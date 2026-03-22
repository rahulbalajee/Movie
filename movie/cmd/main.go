package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/movie/internal/controller/movie"
	metadatagateway "github.com/rahulbalajee/Movie/movie/internal/gateway/metadata/http"
	ratinggateway "github.com/rahulbalajee/Movie/movie/internal/gateway/rating/http"
	httphandler "github.com/rahulbalajee/Movie/movie/internal/handler/http"
)

var (
	port                = ":8083"
	metadataGatewayAddr = "http://localhost:8081"
	ratingGatewayAddr   = "http://localhost:8082"
)

func main() {
	log.Println("Starting the movie service")

	metadataGateway := metadatagateway.New(metadataGatewayAddr)
	ratingGateway := ratinggateway.New(ratingGatewayAddr)

	crtl := movie.NewController(ratingGateway, metadataGateway)
	h := httphandler.NewHandler(crtl)

	mux := http.NewServeMux()
	mux.Handle("GET /movie", http.HandlerFunc(h.GetMovieDetails))

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
