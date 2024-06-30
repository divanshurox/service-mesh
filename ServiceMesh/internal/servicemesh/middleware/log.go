package middleware

import (
	"log"
	"net/http"
)

func AddLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Making a request on %s", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
