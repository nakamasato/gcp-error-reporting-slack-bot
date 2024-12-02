package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
)

type Webhook struct {
    Version       string         `json:"version"`
    Subject       string         `json:"subject"`
    GroupInfo     GroupInfo      `json:"group_info"`
    ExceptionInfo ExceptionInfo  `json:"exception_info"`
    EventInfo     EventInfo      `json:"event_info"`
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

var projectChannelMap = map[string]string{
    "project_id_1": "C055K6NALJK", // Replace with your actual Slack channel IDs
    "project_id_2": "C055K6NALJK",
}

var defaultChannel = "C020D7N58G0" // Replace with your default Slack channel ID

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    log.Println("Server is listening on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
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

    // Proceed with processing the webhook
    processWebhook(webhook)
    w.WriteHeader(http.StatusOK)
}

func processWebhook(webhook Webhook) {
    // Determine the Slack channel
    channelID, ok := projectChannelMap[webhook.GroupInfo.ProjectID]
    if !ok {
        channelID = defaultChannel
    }

    // Build the Slack message
    payload := buildSlackMessage(webhook, channelID)

    // Send the message to Slack
    slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
    if slackBotToken == "" {
        log.Println("SLACK_BOT_TOKEN environment variable is not set")
        return
    }

    err := sendSlackMessage(slackBotToken, payload)
    if err != nil {
        log.Printf("Error sending message to Slack: %v\n", err)
    } else {
        log.Println("Message sent to Slack successfully.")
    }
}

func buildSlackMessage(webhook Webhook, channelID string) map[string]interface{} {
    blocks := []map[string]interface{}{
        {
            "type": "section",
            "text": map[string]interface{}{
                "type": "mrkdwn",
                "text": fmt.Sprintf("*[Alert]* New error reported in service: *%s*", webhook.EventInfo.Service),
            },
        },
        {
            "type": "section",
            "fields": []map[string]interface{}{
                {
                    "type": "mrkdwn",
                    "text": fmt.Sprintf("*Project ID*\n%s", webhook.GroupInfo.ProjectID),
                },
                {
                    "type": "mrkdwn",
                    "text": fmt.Sprintf("*Version*\n%s", webhook.EventInfo.Version),
                },
            },
        },
        {
            "type": "section",
            "text": map[string]interface{}{
                "type": "mrkdwn",
                "text": fmt.Sprintf("*Error Message*\n%s", webhook.ExceptionInfo.Message),
            },
        },
        {
            "type": "actions",
            "elements": []map[string]interface{}{
                {
                    "type": "button",
                    "text": map[string]interface{}{
                        "type": "plain_text",
                        "text": "View Details",
                    },
                    "url": webhook.GroupInfo.DetailLink,
                },
            },
        },
    }

    payload := map[string]interface{}{
        "channel": channelID,
        "blocks":  blocks,
    }

    return payload
}

func sendSlackMessage(token string, payload map[string]interface{}) error {
    url := "https://slack.com/api/chat.postMessage"

    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    var respData map[string]interface{}
    err = json.Unmarshal(body, &respData)
    if err != nil {
        return err
    }

    if !respData["ok"].(bool) {
        return fmt.Errorf("Slack API error: %v", respData["error"])
    }

    return nil
}
