package nimschatwidget

import "net/http"

// Handler returns an http.Handler that serves the widget JS endpoint.
func Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /widget", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Write([]byte(widgetJS))
	})

	mux.HandleFunc("OPTIONS /widget", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
