package webmonitor

import (
	"compress/gzip"
	_ "embed"
	"io"
	"net/http"
	"strings"
)

//go:generate npm install
//go:embed node_modules/alpinejs/dist/cdn.min.js
var alpinejs []byte

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Wrap a handler function to support gzip compression
func gzipHandlerFunc(towrap http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if gzip is supported
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Not supported
			towrap(w, r)
			return
		} else {
			// Supported
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			gzr := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
			towrap(gzr, r)
		}

	})
}
