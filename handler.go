package nimschatwidget

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// NimInfo describes an available nim for the widget dropdown.
type NimInfo struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// DefaultNims is the standard list of NimsForest nims.
var DefaultNims = []NimInfo{
	{Name: "neo", Role: "Technical lead"},
	{Name: "nimble", Role: "General assistant"},
	{Name: "napoleon", Role: "Strategy"},
	{Name: "nectar", Role: "Marketing"},
	{Name: "nudge", Role: "Behavioral design"},
	{Name: "nostradamus", Role: "Predictions"},
	{Name: "numbers", Role: "Finance"},
	{Name: "nebula", Role: "Creative"},
	{Name: "nightwatch", Role: "Security"},
	{Name: "nova", Role: "Innovation"},
	{Name: "narcissus", Role: "Self-reflection"},
	{Name: "navigator", Role: "Direction"},
	{Name: "nebucha", Role: "History"},
	{Name: "nefertiti", Role: "Leadership"},
	{Name: "neuron", Role: "Learning"},
	{Name: "next", Role: "Future planning"},
	{Name: "nexus", Role: "Connections"},
	{Name: "niall", Role: "Storytelling"},
	{Name: "nitty", Role: "Details"},
	{Name: "noble", Role: "Ethics"},
	{Name: "notar", Role: "Legal"},
	{Name: "nourish", Role: "Wellbeing"},
	{Name: "now", Role: "Mindfulness"},
	{Name: "nuclear", Role: "Energy"},
	{Name: "nurture", Role: "Growth"},
	{Name: "nimspeaker", Role: "Communication"},
}

// Handler returns an http.Handler for the nimschatwidget endpoints.
// Routes (relative to mount point):
//   - POST /send         — send a message to a nim
//   - GET  /nims         — list available nims
//   - GET  /events       — SSE stream for a session
//   - GET  /widget       — serve the embedded chat widget JS/CSS/HTML
func Handler(source *Source, songbird *Songbird) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /send", handleSend(source))
	mux.HandleFunc("GET /nims", handleNims)
	mux.HandleFunc("GET /events", handleEvents(songbird))
	mux.HandleFunc("GET /widget", handleWidget)

	return cors(mux)
}

// cors wraps a handler with permissive CORS headers for cross-origin embeds.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleSend(source *Source) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SessionID string `json:"session_id"`
			TargetNim string `json:"target_nim"`
			Text      string `json:"text"`
			Context   string `json:"context"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.SessionID == "" || req.Text == "" {
			http.Error(w, "session_id and text are required", http.StatusBadRequest)
			return
		}

		if req.TargetNim == "" {
			req.TargetNim = "nimble"
		}

		if err := source.Send(req.SessionID, req.TargetNim, req.Text, req.Context); err != nil {
			log.Printf("[ChatWidget] send error: %v", err)
			http.Error(w, "failed to send message", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func handleNims(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DefaultNims)
}

func handleEvents(songbird *Songbird) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session")
		if sessionID == "" {
			http.Error(w, "session parameter required", http.StatusBadRequest)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		ch := songbird.Listen(sessionID)
		defer songbird.Unlisten(sessionID, ch)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				data, _ := json.Marshal(msg)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
