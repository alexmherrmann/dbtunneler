package webmonitor

// -go:embed dashboard.html
// var indexHTML []byte

// -go:embed assets/*
// var assets embed.FS

// type embeddedServer struct {
// 	var assets embed.FS
// 	var indexHTML []byte
// }

// func (e *embeddedServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// w.Header().Set("Content-Type", "text/html")

// if r.URL.Path == "/" || r.URL.Path == "/index.html" {
// 	w.WriteHeader(http.StatusOK)
// 	_, err := w.Write([]byte(indexHTML))
// 	if err != nil {
// 		log.Println("Error writing index.html: ", err)
// 	}
// } else {
// 	// Serve from assets
// 	assetPath := r.URL.Path[1:]

// 	opener, err := assets.Open(assetPath)
// 	if err == fs.ErrNotExist {
// 		w.WriteHeader(http.StatusNotFound)
// 		log.Printf("Asset %s not found\n", assetPath)
// 		return
// 	} else if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		log.Println("Error opening asset: ", err)
// 		return
// 	} else {
// 		defer opener.Close()
// 		w.WriteHeader(http.StatusOK)

// 		_, err := io.Copy(w, opener)
// 		if err != nil {
// 			log.Println("Error copying asset: ", err)
// 		}

// 	}

// }
