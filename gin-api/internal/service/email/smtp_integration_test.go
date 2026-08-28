package email

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testSMTPHost   = "127.0.0.1"
	testSMTPPort   = 1025
	testMailpitURL = "http://127.0.0.1:8025"
)

func TestSMTPSender_SendRegistrationCode_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SMTP integration test in short mode")
	}

	sender := NewSMTPSender(SMTPConfig{
		Host: testSMTPHost,
		Port: testSMTPPort,
		From: "no-reply@sviper.local",
	})

	const (
		recipient = "integration@example.com"
		code      = "481293"
	)

	err := sender.SendRegistrationCode(
		context.Background(),
		recipient,
		code,
	)
	require.NoError(t, err)

	message := waitForMailpitMessage(
		t,
		recipient,
		"Verify your S-VIPER account",
	)

	require.Contains(t, message, code)
	require.Contains(t, message, "The code expires in 10 minutes.")
	require.Contains(t, message, "S-VIPER")
}

func waitForMailpitMessage(
	t *testing.T,
	recipient string,
	subject string,
) string {
	t.Helper()

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	query := url.Values{}
	query.Set("query", recipient)

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		requestURL := fmt.Sprintf(
			"%s/api/v1/search?%s",
			testMailpitURL,
			query.Encode(),
		)

		resp, err := client.Get(requestURL)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()

			if readErr == nil && resp.StatusCode == http.StatusOK {
				var result struct {
					Messages []struct {
						ID      string `json:"ID"`
						Subject string `json:"Subject"`
					} `json:"messages"`
				}

				if json.Unmarshal(body, &result) == nil {
					for _, message := range result.Messages {
						if message.Subject != subject {
							continue
						}

						return getMailpitMessageBody(t, client, message.ID)
					}
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf(
		"message for recipient %q with subject %q was not received by Mailpit",
		recipient,
		subject,
	)

	return ""
}

func getMailpitMessageBody(
	t *testing.T,
	client *http.Client,
	messageID string,
) string {
	t.Helper()

	requestURL := fmt.Sprintf(
		"%s/api/v1/message/%s",
		testMailpitURL,
		messageID,
	)

	resp, err := client.Get(requestURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var message struct {
		Text string `json:"Text"`
		HTML string `json:"HTML"`
	}

	require.NoError(t, json.Unmarshal(body, &message))

	return strings.TrimSpace(message.Text + "\n" + message.HTML)
}
