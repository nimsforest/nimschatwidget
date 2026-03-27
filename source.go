package nimschatwidget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Source publishes chat messages to the forest via the webhook endpoint.
type Source struct {
	webhookURL string
	appName    string
	client     *http.Client
}

// NewSource creates a new Source that POSTs chat messages to the forest webhook.
// webhookURL is the full URL to the chatwidget webhook (e.g. "http://127.0.0.1:8081/webhooks/chatwidget").
// appName identifies the embedding application (e.g., "nimschatwidget").
func NewSource(webhookURL, appName string) *Source {
	return &Source{
		webhookURL: webhookURL,
		appName:    appName,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// chatMessage is the payload POSTed to the forest webhook.
type chatMessage struct {
	SessionID    string `json:"session_id"`
	TargetNim    string `json:"target_nim"`
	UserName     string `json:"user_name"`
	Text         string `json:"text"`
	ReplySubject string `json:"reply_subject"`
	Context      string `json:"context,omitempty"`
	Timestamp    string `json:"timestamp"`
}

// Send publishes a chat message to the forest webhook.
// The reply_subject is set to song.nimschatwidget.{sessionID} so the
// forest pipeline routes responses back to this widget's Songbird.
func (s *Source) Send(sessionID, targetNim, text, context string) error {
	msg := chatMessage{
		SessionID:    sessionID,
		TargetNim:    targetNim,
		UserName:     s.appName,
		Text:         text,
		ReplySubject: "song.nimschatwidget." + sessionID,
		Context:      context,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("post to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("[Source] Published to webhook (session=%s, nim=%s)", sessionID, targetNim)
	return nil
}
