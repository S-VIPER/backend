//go:build integration

package email_test

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
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type SMTPSenderIntegrationTestSuite struct {
	suite.Suite

	ctx       context.Context
	container testcontainers.Container

	apiURL     string
	httpClient *http.Client

	sender *SMTPSender
}

func (s *SMTPSenderIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "axllent/mailpit:v1.21",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForListeningPort("8025/tcp"),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)
	s.container = container

	host, err := container.Host(s.ctx)
	require.NoError(s.T(), err)

	smtpPort, err := container.MappedPort(s.ctx, "1025/tcp")
	require.NoError(s.T(), err)

	apiPort, err := container.MappedPort(s.ctx, "8025/tcp")
	require.NoError(s.T(), err)

	s.apiURL = fmt.Sprintf("http://%s:%s", host, apiPort.Port())
	s.httpClient = &http.Client{Timeout: 2 * time.Second}

	s.sender = NewSMTPSender(SMTPConfig{
		Host: host,
		Port: smtpPort.Int(),
		From: "no-reply@sviper.local",
	})
}

func (s *SMTPSenderIntegrationTestSuite) TearDownSuite() {
	require.NoError(s.T(), s.container.Terminate(s.ctx))
}

func (s *SMTPSenderIntegrationTestSuite) TestSendRegistrationCode() {
	const (
		recipient = "integration@example.com"
		code      = "481293"
	)

	err := s.sender.SendRegistrationCode(s.ctx, recipient, code)
	s.Require().NoError(err)

	message := s.waitForMailpitMessage(recipient, "Verify your S-VIPER account")

	s.Require().Contains(message, code)
	s.Require().Contains(message, "The code expires in 10 minutes.")
	s.Require().Contains(message, "S-VIPER")
}

func (s *SMTPSenderIntegrationTestSuite) waitForMailpitMessage(recipient, subject string) string {
	s.T().Helper()

	query := url.Values{}
	query.Set("query", recipient)

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		requestURL := fmt.Sprintf("%s/api/v1/search?%s", s.apiURL, query.Encode())

		resp, err := s.httpClient.Get(requestURL)
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
					for _, m := range result.Messages {
						if m.Subject != subject {
							continue
						}
						return s.getMailpitMessageBody(m.ID)
					}
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	s.T().Fatalf("message for recipient %q with subject %q was not received by Mailpit", recipient, subject)
	return ""
}

func (s *SMTPSenderIntegrationTestSuite) getMailpitMessageBody(messageID string) string {
	s.T().Helper()

	resp, err := s.httpClient.Get(fmt.Sprintf("%s/api/v1/message/%s", s.apiURL, messageID))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	var message struct {
		Text string `json:"Text"`
		HTML string `json:"HTML"`
	}
	s.Require().NoError(json.Unmarshal(body, &message))

	return strings.TrimSpace(message.Text + "\n" + message.HTML)
}

func TestSMTPSenderIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SMTP integration test in short mode")
	}
	suite.Run(t, new(SMTPSenderIntegrationTestSuite))
}
