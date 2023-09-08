package webmonitor

import (
	"encoding/json"
	"net/http"
)

func (w *WebMonitor) GetMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/alpinejs.js", gzipHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write(alpinejs)
	}))

	// Return the index.html file
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler := gzipHandlerFunc(http.FileServer(http.Dir("public")).ServeHTTP)

		handler.ServeHTTP(w, r)

	})

	mux.HandleFunc("/api/state", func(writer http.ResponseWriter, r *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		unclaim := w.claim()
		bytes, err := json.MarshalIndent(w, "", "  ")
		unclaim()

		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write(bytes)
	})

	return mux
}
