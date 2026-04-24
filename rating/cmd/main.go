package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rahulbalajee/Movie/gen"
	"github.com/rahulbalajee/Movie/pkg/discovery"
	"github.com/rahulbalajee/Movie/pkg/discovery/consul"
	"github.com/rahulbalajee/Movie/rating/internal/controller/rating"
	grpchandler "github.com/rahulbalajee/Movie/rating/internal/handler/grpc"
	"github.com/rahulbalajee/Movie/rating/internal/ingester/kafka"
	"github.com/rahulbalajee/Movie/rating/internal/repository/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func main() {
	var port, serviceName, consulAddr, kafkaAddr string
	flag.StringVar(&port, "port", "8082", "API handler port")
	flag.StringVar(&serviceName, "service-name", "rating", "service name")
	flag.StringVar(&consulAddr, "consul-addr", "localhost:8500", "consul address")
	flag.StringVar(&kafkaAddr, "kafka-addr", "localhost:9092", "kafka bootstrap server")
	flag.Parse()

	log.Printf("Starting the rating service on port %s", port)

	registry, err := consul.NewRegistry(consulAddr)
	if err != nil {
		log.Fatalf("error creating consul registry: %v", err)
	}

	instanceId := discovery.GenerateInstanceId(serviceName)

	if err := registry.Register(context.Background(), instanceId, serviceName, fmt.Sprintf("localhost:%s", port)); err != nil {
		log.Fatalf("error registering service with consul: %v", err)
	}

	healthCtx, cancel := context.WithCancel(context.Background())

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

	repo := memory.NewRepo()

	ingester, err := kafka.NewIngester(kafkaAddr, "rating", "ratings")
	if err != nil {
		log.Fatalf("failed to initialize ingester: %v", err)
	}

	ctrl := rating.NewController(repo, ingester)

	if err := ctrl.StartIngestion(context.Background()); err != nil {
		log.Fatalf("failed to start ingestion: %v", err)
	}

	h := grpchandler.NewHandler(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
		grpc.MaxSendMsgSize(4*1024*1024),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              2 * time.Minute,
			Timeout:           20 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: false,
		}),
	)
	gen.RegisterRatingServiceServer(srv, h)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Printf("gRPC server stopped with error: %v", err)
			select {
			case quit <- syscall.SIGTERM: // trigger shutdown
			default: // shutdown already in progress
			}
		}
	}()

	<-quit
	signal.Stop(quit)
	log.Println("Shutting down gRPC server...")

	// Not deferred — order matters: stop health checks, deregister, then drain.
	cancel()
	if err := registry.Deregister(context.Background(), instanceId, serviceName); err != nil {
		log.Printf("failed to deregister from consul: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("Graceful shutdown timed out, forcing stop")
		srv.Stop()
	}
}
