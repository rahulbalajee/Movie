package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/metadata/internal/controller/metadata"
	httphandler "github.com/rahulbalajee/Movie/metadata/internal/handler/http"
	"github.com/rahulbalajee/Movie/metadata/internal/repository/memory"
)

var (
	port = ":8081"
)

func main() {
	log.Println("starting the movie metadata service")

	repo := memory.NewRepo()
	ctrl := metadata.NewController(repo)
	h := httphandler.NewHandler(ctrl)

	mux := http.NewServeMux()
	mux.Handle("GET /metadata", http.HandlerFunc(h.GetMetadata))

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
