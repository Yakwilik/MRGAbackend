package scratch

import (
	"net/http"
	"time"
)

var allowedHosts = []string{}
var now = time.Now()

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ApiVersion", now.String())
		if r.Header.Get("Origin") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		}
		//if allowedOrigin(r.Header.Get("Origin")) {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, ResponseType")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		//}
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}
