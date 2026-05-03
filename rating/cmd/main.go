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
	"github.com/rahulbalajee/Movie/rating/internal/repository/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/yaml.v3"
)

func main() {
	var serviceName, configPath string
	flag.StringVar(&serviceName, "service-name", "rating", "service name")
	flag.StringVar(&configPath, "config", "rating/configs/default.yaml", "Path to config file")
	flag.Parse()

	log.Println("Starting the rating service")

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("failed to open config file: %v", err)
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("failed to decode config file: %v", err)
	}

	// Env vars override file values for environment-specific / sensitive settings,
	// so prod can inject real creds via Secrets without committing them to YAML.
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("CONSUL_ADDRESS"); v != "" {
		cfg.ServiceDiscovery.Consul.Address = v
	}
	if v := os.Getenv("KAFKA_ADDRESS"); v != "" {
		cfg.Ingester.Kafka.Address = v
	}

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		log.Fatalf("error creating consul registry: %v", err)
	}

	instanceId := discovery.GenerateInstanceId(serviceName)
	advertiseAddr := fmt.Sprintf("%s:%d", cfg.API.AdvertiseHost, cfg.API.Port)
	if err := registry.Register(context.Background(), instanceId, serviceName, advertiseAddr); err != nil {
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

	repo, err := mysql.NewRepository(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	ingester, err := kafka.NewIngester(cfg.Ingester.Kafka.Address, "rating", "ratings")
	if err != nil {
		log.Fatalf("failed to initialize ingester: %v", err)
	}

	ctrl := rating.NewController(repo, ingester)

	go func() {
		if err := ctrl.StartIngestion(context.Background()); err != nil {
			log.Printf("ingestion stopped: %v", err)
		}
	}()

	h := grpchandler.NewHandler(ctrl)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port))
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
