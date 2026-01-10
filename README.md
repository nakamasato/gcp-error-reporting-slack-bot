# GCP Error Reporting Slack Bot

<img src="image.png" alt="drawing" width="500"/>

A Go application that receives Google Cloud Platform (GCP) Error Reporting webhooks and sends formatted alerts to specified Slack channels. This bot helps teams stay informed about errors occurring in their GCP projects by delivering real-time notifications directly to Slack.

```mermaid
sequenceDiagram
    participant GCP as GCP Error Reporting
    participant App as gcp-error-reporting-slack-bot
    participant Slack as Slack

    GCP->>App: HTTP POST Event via Webhook
    App->>Slack: Send Message via Slack API
    Slack-->>App: Acknowledge Message Receipt
    App-->>GCP: HTTP 200 OK
```

## Features

- Webhook Receiver: Accepts GCP Error Reporting webhooks with detailed error information.
- Slack Integration: Sends error notifications to Slack channels using rich formatting.
- Project-Based Channel Mapping: Configurable mapping of GCP project IDs to Slack channels.
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

```hcl
resource "google_monitoring_notification_channel" "gcp_error_reporting_slack_bot" {
  display_name = "gcp-error-reporting-slack-bot"
  type         = "webhook_basicauth"
  labels = {
    username = "gcp-error-reporting"
    url      = "https://<cloud run service name of gcp_error_reporting_slack_bot>-<project number>.asia-northeast1.run.app/webhook"
  }

  sensitive_labels {
    password = "xxxx" # only valid for webhook_basicauth
  }
}
```

## Environment Variables

- `SLACK_BOT_TOKEN`: `xoxb-your-slack-bot-token`
- `SLACK_CHANNEL_MAP`: `project_id_1:C01ABCD2EFG,project_id_2:C02HIJK3LMN`
- `SLACK_DEFAULT_CHANNEL`: default slack channel id (e.g. `C03OPQR4STU`)
- `BASIC_AUTH_USERNAME`: basic auth username
- `BASIC_AUTH_PASSWORD`: basic auth password

## Testing

```
go run main.go
```

```
curl -X POST -H "Content-Type: application/json" -u your-username:your-password -d '{
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

> [!NOTE]
> Or you can add the Header `Authorization: Basic <base64 of <your-username>:<your-password>>`
> ```
> echo -n 'your-username:your-password' | base64
> eW91ci11c2VybmFtZTp5b3VyLXBhc3N3b3Jk
> ```

## References

- [Deliver exception messages through Slack and Webhooks for fast resolution](https://cloud.google.com/blog/products/devops-sre/use-slack-and-webhooks-for-notifications)
- [Slack API](https://api.slack.com/methods)
