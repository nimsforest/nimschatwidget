package nimschatwidget

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/nimsforest/nimsforest2/pkg/nim"
)

const songbirdPattern = "song.nimschatwidget.>"

// Message represents a chat response from a nim.
type Message struct {
	Text    string `json:"text"`
	Source  string `json:"source"`
	IsError bool   `json:"is_error,omitempty"`
}

// Songbird catches nim responses and delivers them to SSE listeners.
type Songbird struct {
	wind      *nim.Wind
	mu        sync.RWMutex
	listeners map[string]map[chan Message]struct{} // sessionID → SSE channels
}

// NewSongbird creates a new Songbird that catches song.nimschatwidget.> responses.
func NewSongbird(wind *nim.Wind) *Songbird {
	return &Songbird{
		wind:      wind,
		listeners: make(map[string]map[chan Message]struct{}),
	}
}

// Start subscribes to song.nimschatwidget.> and routes responses to listeners.
func (s *Songbird) Start() error {
	_, err := s.wind.Catch(songbirdPattern, func(leaf nim.Leaf) {
		// Extract session_id from subject: song.nimschatwidget.{session_id}
		parts := strings.Split(leaf.Subject, ".")
		if len(parts) < 3 {
			log.Printf("[ChatWidget Songbird] unexpected subject: %s", leaf.Subject)
			return
		}
		sessionID := strings.Join(parts[2:], ".")

		if strings.HasPrefix(sessionID, "{") {
			log.Printf("[ChatWidget Songbird] unresolved placeholder: %s", leaf.Subject)
			return
		}

		var response struct {
			Text     string `json:"text"`
			Response string `json:"response"`
			Error    bool   `json:"error"`
		}
		if err := json.Unmarshal(leaf.Data, &response); err != nil {
			log.Printf("[ChatWidget Songbird] parse error: %v", err)
			return
		}

		text := response.Text
		if text == "" {
			text = response.Response
		}

		nimName := strings.TrimPrefix(leaf.Source, "nim:")

		msg := Message{
			Text:    text,
			Source:  nimName,
			IsError: response.Error,
		}

		s.chirp(sessionID, msg)
	})
	return err
}

// chirp sends a message to all listeners for a session.
func (s *Songbird) chirp(sessionID string, msg Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channels, ok := s.listeners[sessionID]
	if !ok {
		return
	}

	for ch := range channels {
		select {
		case ch <- msg:
		default:
			// drop if subscriber is slow
		}
	}
}

// Listen registers an SSE connection for a session. Returns a buffered message channel.
func (s *Songbird) Listen(sessionID string) chan Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan Message, 32)
	if s.listeners[sessionID] == nil {
		s.listeners[sessionID] = make(map[chan Message]struct{})
	}
	s.listeners[sessionID][ch] = struct{}{}
	return ch
}

// Unlisten removes an SSE connection for a session.
func (s *Songbird) Unlisten(sessionID string, ch chan Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if channels, ok := s.listeners[sessionID]; ok {
		delete(channels, ch)
		if len(channels) == 0 {
			delete(s.listeners, sessionID)
		}
	}
	close(ch)
}
