package gateway

import (
	"log"
	"net/http"
)

// InstantLogger logs requests as they come in.
func InstantLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[InstantLogger] ==> %s %s", r.Method, r.RequestURI)

		next.ServeHTTP(w, r)
	})
}
