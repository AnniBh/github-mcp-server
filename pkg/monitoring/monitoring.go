package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/mark3labs/mcp-go/mcp"
)

const webhookURL = "https://annibhaskerpce.app.n8n.cloud/webhook/ed3220ad-38b7-46ea-9cbe-3de54fa3d4d7"

type UserDetails struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type FetchUserDetailsFunc func(ctx context.Context) (UserDetails, error)

type WebhookPayload struct {
	Timestamp       string        `json:"timestamp"`
	ToolName        string        `json:"tool_name"`
	UserDetails     UserDetails   `json:"user_details"`
	ResponseData    interface{}   `json:"response_data"`
	ExecutionTimeS  float64       `json:"execution_time_ms"`
}

func WithMonitoring(tool server.ServerTool, fetchUserDetails FetchUserDetailsFunc) server.ServerTool {
	originalExecute := tool.Handler

	tool.Handler = server.ToolHandlerFunc(func(ctx context.Context, input mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()

		var userDetails UserDetails
		if fetchUserDetails != nil {
			var err error
			userDetails, err = fetchUserDetails(ctx)
			if err != nil {
				// Log the error but continue execution
				logrus.WithError(err).Error("failed to fetch user details")
			}
		}

		response, err := originalExecute(ctx, input)
		if err != nil {
			return nil, err
		}

		executionTime := time.Since(start).Seconds()

		payload := WebhookPayload{
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			ToolName:       tool.Tool.Name,
			UserDetails:    userDetails,
			ResponseData:   nil,
			ExecutionTimeS: executionTime,
		}

		go sendToWebhook(payload)

		return response, nil
	})

	return tool
}

func sendToWebhook(payload WebhookPayload) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal webhook payload")
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.WithError(err).Error("failed to create webhook request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("failed to send webhook request")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logrus.WithField("status_code", resp.StatusCode).Error("webhook returned non-200 status")
	}
} 