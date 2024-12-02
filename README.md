# GCP Error Reporting Slack Bot

A Go application that receives Google Cloud Platform (GCP) Error Reporting webhooks and sends formatted alerts to specified Slack channels. This bot helps teams stay informed about errors occurring in their GCP projects by delivering real-time notifications directly to Slack.

## Features

- Webhook Receiver: Accepts GCP Error Reporting webhooks with detailed error information.
- Slack Integration: Sends error notifications to Slack channels using rich formatting.
- Project-Based Channel Mapping: Configurable mapping of GCP project IDs to Slack channels.
- Customizable Alerts: Uses Slack attachments to highlight errors with red color and actionable buttons.
- Environment Variable Configuration: Easily configure settings without changing code.

## Prerequisites

- Go 1.20 or later
- Slack API token
- GCP Error Reporting webhook URL

## Configuration

### 1. Slack App
1. Create a Slack app.
2. Oauth & Permissions
    - channels:read
    - chat:write
    - chat:write.public
3. Obtain the Bot User OAuth Token.
4. Invite the bot to the Slack channels where you want to send notifications.
    ```
    /invite @<bot-name>
    ```

### 2. Deploy

1. Deploy the application to Cloud Run.
2. Set the environment variables.

### 3. Error Reporting

1. Set the webhook URL to the Error Reporting webhook URL.

## Environment Variables

- SLACK_BOT_TOKEN
- SLACK_CHANNEL_MAP
- SLACK_DEFAULT_CHANNEL

Examples
```
export SLACK_BOT_TOKEN="xoxb-your-slack-bot-token"
export PROJECT_CHANNEL_MAP="project_id_1:C01ABCD2EFG,project_id_2:C02HIJK3LMN"
export DEFAULT_CHANNEL_ID="C03OPQR4STU"
```

## Testing

```
go run main.go
```

```
curl -X POST -H "Content-Type: application/json" -d '{
  "version": "1.0",
  "subject": "Error in production",
  "group_info": {
    "project_id": "project_id_1",
    "detail_link": "https://console.cloud.google.com/errors/12345?project=project_id_1"
  },
  "exception_info": {
    "type": "NullReferenceException",
    "message": "Object reference not set to an instance of an object."
  },
  "event_info": {
    "log_message": "An error occurred.",
    "request_method": "GET",
    "request_url": "https://example.com/api/data",
    "referrer": "https://example.com/home",
    "user_agent": "Mozilla/5.0",
    "service": "UserService",
    "version": "v1.2.3",
    "response_status": "500"
  }
}' http://localhost:8080/webhook
```

## References

- [Deliver exception messages through Slack and Webhooks for fast resolution](https://cloud.google.com/blog/products/devops-sre/use-slack-and-webhooks-for-notifications)
- [Slack API](https://api.slack.com/methods)
