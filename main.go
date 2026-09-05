package main

import (
	"log"
	"net/http"

	"featureflags/internal/handlers"
	"featureflags/internal/middleware"
	"featureflags/internal/store"
)

func main() {
	s := store.New()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", handlers.Create(s))
	mux.HandleFunc("GET /flags", handlers.List(s))
	mux.HandleFunc("GET /flags/{key}", handlers.Get(s))
	mux.HandleFunc("PUT /flags/{key}", handlers.Update(s))
	mux.HandleFunc("DELETE /flags/{key}", handlers.Delete(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", handlers.Evaluate(s))
	mux.HandleFunc("GET /healthz", handlers.Healthz())

	handler := middleware.Logging(mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
