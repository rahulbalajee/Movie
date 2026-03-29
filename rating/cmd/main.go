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

func main() {
	var port, serviceName, consulAddr string
	flag.StringVar(&port, "port", "8082", "API handler port")
	flag.StringVar(&serviceName, "service-name", "rating", "service name")
	flag.StringVar(&consulAddr, "consul-addr", "localhost:8500", "consul address")
	flag.Parse()

	log.Printf("Starting the rating service on port %s", port)

	registry, err := consul.NewRegistry(consulAddr)
	if err != nil {
		log.Fatal(err)
	}

	instanceId := discovery.GenerateInstanceId(serviceName)
	ctx := context.Background()
	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%s", port)); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(ctx, instanceId, serviceName); err != nil {
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
		Addr:              fmt.Sprintf(":%s", port),
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
