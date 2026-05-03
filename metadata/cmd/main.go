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
	"github.com/rahulbalajee/Movie/metadata/internal/controller/metadata"
	grpchandler "github.com/rahulbalajee/Movie/metadata/internal/handler/grpc"
	"github.com/rahulbalajee/Movie/metadata/internal/repository/lru"
	"github.com/rahulbalajee/Movie/metadata/internal/repository/mysql"
	"github.com/rahulbalajee/Movie/pkg/discovery"
	"github.com/rahulbalajee/Movie/pkg/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/yaml.v3"
)

func main() {
	var serviceName, configPath string
	flag.StringVar(&serviceName, "service-name", "metadata", "service name")
	flag.StringVar(&configPath, "config", "metadata/configs/default.yaml", "path to config file")
	flag.Parse()

	log.Println("Starting the movie metadata service")

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("error opening config file: %v", err)
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("error decoding config file: %v", err)
	}

	// Env vars override file values for environment-specific / sensitive settings,
	// so prod can inject real creds via Secrets without committing them to YAML.
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("CONSUL_ADDRESS"); v != "" {
		cfg.ServiceDiscovery.Consul.Address = v
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

	// Health reporter — pings Consul every second, exits when healthCtx is cancelled.
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
	// Bounded LRU cache in front of repo: 10K hot entries, 5m TTL so stale
	// metadata self-heals without an explicit invalidation path.
	cache := lru.New(10_000, 5*time.Minute)
	ctrl := metadata.NewController(repo, cache)
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
	gen.RegisterMetadataServiceServer(srv, h)

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

	// Drain in-flight RPCs with a 10s deadline.
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
