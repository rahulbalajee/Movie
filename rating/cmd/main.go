package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rahulbalajee/Movie/pkg/discovery"
	"github.com/rahulbalajee/Movie/pkg/discovery/consul"
	"github.com/rahulbalajee/Movie/rating/internal/controller/rating"
	httphandler "github.com/rahulbalajee/Movie/rating/internal/handler/http"
	"github.com/rahulbalajee/Movie/rating/internal/repository/memory"
)

const (
	serviceName   = "rating"
	consulDevAddr = "localhost:8500"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()

	log.Printf("Starting the rating service on port %d", port)

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

	repo := memory.NewRepo()
	ctrl := rating.NewController(repo)
	h := httphandler.NewHandler(ctrl)

	mux := http.NewServeMux()
	mux.Handle("GET /rating", http.HandlerFunc(h.GetRating))
	mux.Handle("PUT /rating", http.HandlerFunc(h.PutRating))

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
