package nimschatwidget

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nimsforest/nimsforest2/pkg/nim"
)

const riverSubject = "river.chat.widget"

// Source publishes chat messages into the forest river via JetStream.
type Source struct {
	wind    *nim.Wind
	appName string
}

// NewSource creates a new Source that publishes chat messages to the forest river.
// appName identifies the embedding application (e.g., "nimschatwidget").
func NewSource(wind *nim.Wind, appName string) *Source {
	return &Source{
		wind:    wind,
		appName: appName,
	}
}

// chatMessage is the payload published to the river.
type chatMessage struct {
	SessionID    string `json:"session_id"`
	TargetNim    string `json:"target_nim"`
	UserName     string `json:"user_name"`
	Text         string `json:"text"`
	ReplySubject string `json:"reply_subject"`
	Context      string `json:"context,omitempty"`
	Timestamp    string `json:"timestamp"`
}

// riverData wraps the payload for JetStream consumption by the forest river.
type riverData struct {
	Subject   string          `json:"subject"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"ts"`
}

// Send publishes a chat message to the forest river via JetStream.
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

	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	rd := riverData{
		Subject:   riverSubject,
		Data:      json.RawMessage(msgData),
		Timestamp: time.Now(),
	}

	payload, err := json.Marshal(rd)
	if err != nil {
		return fmt.Errorf("marshal river data: %w", err)
	}

	js, err := s.wind.JetStream()
	if err != nil {
		return fmt.Errorf("get jetstream: %w", err)
	}

	if _, err := js.Publish(riverSubject, payload); err != nil {
		return fmt.Errorf("publish to river: %w", err)
	}

	log.Printf("[Source] Published to %s (session=%s, nim=%s)", riverSubject, sessionID, targetNim)
	return nil
}
