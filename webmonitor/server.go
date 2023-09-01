package webmonitor

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
)

//go:embed dashboard.html
var indexHTML []byte

//go:embed assets/*
var assets embed.FS

func (w *WebMonitor) GetMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Return the index.html file
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(indexHTML))
			if err != nil {
				log.Println("Error writing index.html: ", err)
			}
		} else {
			// Serve from assets
			assetPath := r.URL.Path[1:]

			opener, err := assets.Open(assetPath)
			if err == fs.ErrNotExist {
				w.WriteHeader(http.StatusNotFound)
				log.Printf("Asset %s not found\n", assetPath)
				return
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println("Error opening asset: ", err)
				return
			} else {
				defer opener.Close()
				w.WriteHeader(http.StatusOK)

				_, err := io.Copy(w, opener)
				if err != nil {
					log.Println("Error copying asset: ", err)
				}

			}

		}
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
