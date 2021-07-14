package middleware

import (
	"log"
	"net/http"
)

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		println()
		println("----------------------------------------------------------")
		log.Printf("Start: %s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
		log.Printf("Finish: %s %s", r.Method, r.URL.Path)
		println("----------------------------------------------------------")
		println()
	})
}
