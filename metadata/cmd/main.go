package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rahulbalajee/Movie/metadata/internal/controller/metadata"
	httphandler "github.com/rahulbalajee/Movie/metadata/internal/handler/http"
	"github.com/rahulbalajee/Movie/metadata/internal/repository/memory"
	"github.com/rahulbalajee/Movie/pkg/discovery"
	"github.com/rahulbalajee/Movie/pkg/discovery/consul"
)

func main() {
	var port, serviceName, consulAddr string
	flag.StringVar(&port, "port", "8081", "API handler port")
	flag.StringVar(&serviceName, "service-name", "metadata", "service name")
	flag.StringVar(&consulAddr, "consul-addr", "localhost:8500", "consul address")
	flag.Parse()

	log.Printf("Starting the movie metadata service on port %s", port)

	registry, err := consul.NewRegistry(consulAddr)
	if err != nil {
		log.Fatal(err)
	}

	instanceId := discovery.GenerateInstanceId(serviceName)
	ctx := context.Background()

	if err := registry.Register(ctx, instanceId, serviceName, fmt.Sprintf("localhost:%s", port)); err != nil {
		log.Fatal(err)
	}

	healthCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-healthCtx.Done():
				return
			case <-ticker.C:
				if err := registry.ReportHealthyState(healthCtx, instanceId, serviceName); err != nil {
					log.Println("failed to report healthy status", err)
				}
			}
		}
	}()

	defer registry.Deregister(ctx, instanceId, serviceName)

	repo := memory.NewRepo()
	ctrl := metadata.NewController(repo)
	h := httphandler.NewHandler(ctrl)

	mux := http.NewServeMux()
	mux.Handle("GET /metadata", http.HandlerFunc(h.GetMetadata))

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Printf("error starting the server: %v\n", err)
	case sig := <-quit:
		log.Printf("server is shutting down due to %v signal\n", sig)
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown server gracefully: %v\n", err)
			srv.Close()
		}
	}
}
