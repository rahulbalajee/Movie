package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/rating/internal/controller/rating"
	httphandler "github.com/rahulbalajee/Movie/rating/internal/handler/http"
	"github.com/rahulbalajee/Movie/rating/internal/repository/memory"
)

var (
	port = ":8082"
)

func main() {
	fmt.Println("Starting the rating service")

	repo := memory.NewRepo()
	ctrl := rating.NewController(repo)
	h := httphandler.NewHandler(ctrl)

	mux := http.NewServeMux()
	mux.Handle("GET /rating", http.HandlerFunc(h.GetRating))
	mux.Handle("PUT /rating", http.HandlerFunc(h.PutRating))

	srv := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
