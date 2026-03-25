package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/movie/internal/controller/movie"
	metadatagateway "github.com/rahulbalajee/Movie/movie/internal/gateway/metadata/http"
	ratinggateway "github.com/rahulbalajee/Movie/movie/internal/gateway/rating/http"
	httphandler "github.com/rahulbalajee/Movie/movie/internal/handler/http"
	"github.com/rahulbalajee/Movie/pkg/discovery"
	"github.com/rahulbalajee/Movie/pkg/discovery/consul"
)

var (
	serviceName   = "movie"
	consulDevAddr = "localhost:8500"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API handler port")
	flag.Parse()

	log.Printf("Starting the movie service on port %d", port)

	registry, err := consul.NewRegistry(consulDevAddr)
	if err != nil {
		log.Fatal(err)
	}

	instanceId := discovery.GenerateInstanceId(serviceName)
	ctx := context.Background()
	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceId, serviceName); err != nil {
				log.Println("failed to report healthy status", err)
			}
			time.Sleep(time.Second)
		}
	}()

	defer registry.Deregister(ctx, instanceId, serviceName)

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)

	crtl := movie.NewController(ratingGateway, metadataGateway)
	h := httphandler.NewHandler(crtl)

	mux := http.NewServeMux()
	mux.Handle("GET /movie", http.HandlerFunc(h.GetMovieDetails))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
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
