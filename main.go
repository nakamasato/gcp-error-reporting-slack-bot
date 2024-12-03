package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

type Webhook struct {
	Version       string        `json:"version"`
	Subject       string        `json:"subject"`
	GroupInfo     GroupInfo     `json:"group_info"`
	ExceptionInfo ExceptionInfo `json:"exception_info"`
	EventInfo     EventInfo     `json:"event_info"`
}

type GroupInfo struct {
	ProjectID  string `json:"project_id"`
	DetailLink string `json:"detail_link"`
}

type ExceptionInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type EventInfo struct {
	LogMessage     string `json:"log_message"`
	RequestMethod  string `json:"request_method"`
	RequestURL     string `json:"request_url"`
	Referrer       string `json:"referrer"`
	UserAgent      string `json:"user_agent"`
	Service        string `json:"service"`
	Version        string `json:"version"`
	ResponseStatus string `json:"response_status"`
}

var projectChannelMap map[string]string
var defaultChannel string

func main() {
	// 環境変数からプロジェクトとチャンネルのマッピングをロード
	err := loadProjectChannelMap()
	if err != nil {
		log.Fatalf("Error loading project channel map: %v", err)
	}

	// デフォルトチャンネルを環境変数から取得
	defaultChannel = os.Getenv("DEFAULT_CHANNEL_ID")
	if defaultChannel == "" {
		log.Println("DEFAULT_CHANNEL_ID environment variable is not set")
		// 必要に応じてデフォルト値を設定
	}

	http.HandleFunc("/webhook", webhookHandler)
	log.Println("Server is listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadProjectChannelMap() error {
	projectChannelMap = make(map[string]string)

	mappingStr := os.Getenv("PROJECT_CHANNEL_MAP")
	if mappingStr == "" {
		return fmt.Errorf("PROJECT_CHANNEL_MAP environment variable is not set")
	}

	pairs := strings.Split(mappingStr, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) != 2 {
			return fmt.Errorf("Invalid mapping pair: %s", pair)
		}
		projectID := strings.TrimSpace(kv[0])
		channelID := strings.TrimSpace(kv[1])
		projectChannelMap[projectID] = channelID
	}

	return nil
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Basic認証のチェック
	username := os.Getenv("BASIC_AUTH_USERNAME")
	password := os.Getenv("BASIC_AUTH_PASSWORD")
	if username == "" || password == "" {
		log.Println("Basic auth credentials are not set in environment variables")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	providedUsername, providedPassword, ok := r.BasicAuth()
	if !ok || providedUsername != username || providedPassword != password {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhook Webhook
	err := json.NewDecoder(r.Body).Decode(&webhook)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Webhookを処理
	processWebhook(webhook)
	w.WriteHeader(http.StatusOK)
}

func processWebhook(webhook Webhook) {
	// Slackチャンネルを決定
	channelID, ok := projectChannelMap[webhook.GroupInfo.ProjectID]
	if !ok {
		channelID = defaultChannel
	}

	// Slackクライアントを作成
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackBotToken == "" {
		log.Println("SLACK_BOT_TOKEN environment variable is not set")
		return
	}

	client := slack.New(slackBotToken)

	// Slackメッセージを構築
	msgOptions := buildSlackMessage(webhook)

	// メッセージをSlackに送信
	channelID, timestamp, err := client.PostMessage(
		channelID,
		msgOptions...,
	)
	if err != nil {
		log.Printf("Error sending message to Slack: %v\n", err)
	} else {
		log.Printf("Message sent to Slack successfully. Channel: %s, Timestamp: %s\n", channelID, timestamp)
	}
}

func buildSlackMessage(webhook Webhook) []slack.MsgOption {
	attachment := slack.Attachment{
		Color:     "#ff0000", // 赤色を指定
		Title:     fmt.Sprintf("[Alert] New error reported in service: %s", webhook.EventInfo.Service),
		TitleLink: webhook.GroupInfo.DetailLink,
		Fields: []slack.AttachmentField{
			{
				Title: "Project ID",
				Value: webhook.GroupInfo.ProjectID,
				Short: true,
			},
			{
				Title: "Version",
				Value: webhook.EventInfo.Version,
				Short: true,
			},
			{
				Title: "Error Message",
				Value: webhook.ExceptionInfo.Message,
			},
		},
		Actions: []slack.AttachmentAction{
			{
				Type:  "button",
				Text:  "View Details",
				URL:   webhook.GroupInfo.DetailLink,
				Style: "danger", // ボタンを赤くする
			},
		},
	}

	return []slack.MsgOption{
		slack.MsgOptionAttachments(attachment),
	}
}
