# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

The `Makefile` is the primary entry point and wraps the most common workflows:

```bash
# Run a single service locally (each picks up its own configs/default.yaml)
make run-metadata        # or: run-rating, run-movie, run-all
make run-rating-producer # seeds Kafka from cmd/ratingproducer/ratingsdata.json

# Docker images (one per service; Dockerfile lives next to cmd/)
make build-metadata      # or: build-rating, build-movie, build-all, clean

# Local infra in containers
make consul-up           # hashicorp/consul -dev on :8500
make rating-producer-up  # docker-compose for Kafka + Zookeeper

# Kubernetes (application deployments only — infra is installed via Helm; see k8s/README.md)
make k8s-apply           # or: k8s-delete, k8s-status
```

Lower-level commands when the Makefile doesn't fit:

```bash
# Direct go build of all binaries (used by Dockerfiles)
go build ./metadata/cmd ./rating/cmd ./movie/cmd ./cmd/ratingproducer ./cmd/sizecompare

# Tests
go test ./...
go test ./cmd/sizecompare/... -bench=.            # serialization size benchmarks
go test ./path/to/package -run TestName           # single test

# Regenerate protobuf code (requires protoc, protoc-gen-go, protoc-gen-go-grpc)
protoc -I api/ api/movie.proto --go_out=gen/ --go_opt=paths=source_relative --go-grpc_out=gen/ --go-grpc_opt=paths=source_relative

# Apply MySQL schema (DSN defaults to root:password@/movieexample in configs/default.yaml)
mysql -u root -p movieexample < schema/schema.sql
```

## Architecture

Three microservices communicating via gRPC and HTTP, with Consul-based service discovery:

- **metadata** (gRPC, port 8081) — stores movie metadata (title, description, director)
- **rating** (gRPC, port 8082) — stores and aggregates user ratings
- **movie** (gRPC, port 8083) — aggregation service that calls metadata + rating to build complete movie details

All services register with Consul on startup, report health via TTL checks every 1s, and deregister on graceful shutdown (SIGINT/SIGTERM). The Consul registration also sets `DeregisterCriticalServiceAfter: 1m` so SIGKILL'd or crashed instances are reaped automatically — Consul rounds anything below 1m up, so don't lower it expecting faster cleanup.

### Internal structure per service

Each service follows the same layered layout: `cmd/main.go` → `internal/handler/` → `internal/controller/` → `internal/repository/`. The movie service replaces repository with `internal/gateway/` since it fetches data from the other two services rather than storing its own.

### Configuration

Each service reads `<service>/configs/default.yaml` at startup (path overridable via `-config` flag). The `cmd/config.go` next to each `main.go` defines the schema (api host/port/advertiseHost, consul address, DB DSN, kafka address for rating). After loading the YAML, `main.go` overlays env vars `DB_DSN` and `CONSUL_ADDRESS` so secrets and per-environment endpoints can come from k8s Secrets/ConfigMaps without committing them. `api.host` is the bind address (blank = all interfaces), `api.advertiseHost` is what gets registered with Consul — these diverge in k8s where pods bind locally but advertise their service DNS name.

### Persistence

Both metadata and rating persist to MySQL (`internal/repository/mysql/`); the in-memory repos under `internal/repository/memory/` remain for tests. Schema lives in `schema/schema.sql` (single `movieexample` database with `movies` and `ratings` tables). DSN comes from `configs/default.yaml` (default `root:password@/movieexample`) or the `DB_DSN` env var override.

### Metadata caching (non-obvious)

`metadata/internal/repository/lru/` wraps an `expirable.LRU` (size 10K, TTL 5m) and implements the same `Get/Put/Delete` interface as the MySQL repo. The metadata controller (`metadata/internal/controller/metadata/controller.go`) accepts both as `metadataRepository` and layers them: cache-first reads, write-through invalidation (delete-on-Put, not write-on-Put), and a `singleflight.Group` that collapses concurrent misses for the same id into one downstream query. `Put` also calls `sf.Forget(id)` to prevent a pre-write in-flight Get from repopulating the cache with stale data.

### Rating ingestion (Kafka)

The rating service has two write paths: synchronous gRPC `PutRating` calls, and an asynchronous Kafka consumer (`rating/internal/ingester/kafka/`) that subscribes to the `ratings` topic and feeds events into `controller.StartIngestion`, launched as a goroutine from `rating/cmd/main.go`. The ingester uses manual offset storage (`enable.auto.offset.store=false`) and only stores the offset after the event is delivered to the channel, so a crash mid-processing replays rather than drops. The `cmd/ratingproducer/` tool reads `ratingsdata.json` and produces to the same topic — useful for seeding.

### Shared packages

`pkg/discovery/` defines the `Registry` interface (Register, Deregister, ServiceAddresses, ReportHealthyState) with two implementations: `consul/` (production, requires Consul at localhost:8500) and `memory/` (in-process, for testing). `internal/grpcutil/` centralizes the discovery → gRPC dial dance used by every gateway.

### Protobuf / gRPC

Proto definitions live in `api/movie.proto`. Generated Go code lives in `gen/` (`movie.pb.go`, `movie_grpc.pb.go`). All three services serve gRPC. Legacy HTTP handlers and gateways still exist under `internal/handler/http/` and `movie/internal/gateway/.../http/` but are no longer wired into `cmd/main.go`.

### Service communication flow

```
movie service --gRPC--> metadata service
movie service --gRPC--> rating service
```

### Deployment (Kubernetes)

Each service has a `Dockerfile` and a `kubernetes-deployment.yaml` next to its `cmd/`. Application Deployments+Services are managed via `make k8s-apply`/`k8s-delete`. Infrastructure (Consul, MySQL, Kafka) is installed manually via Helm — see `k8s/README.md` for exact commands and the `kafka-values.yaml`/`mysql-values.yaml` overrides (Bitnami images are pinned to the `bitnamilegacy/*` paths and MySQL to 8.4 LTS for init-script compatibility — don't bump these without reading that README).

## Key conventions

- Module path: `github.com/rahulbalajee/Movie`
- Production repos are MySQL; the `memory` repos remain for tests. The cache layer reuses the same repo interface, which is why the metadata controller can accept both.
- Controllers accept interfaces, not concrete types (repository pattern + dependency injection)
- Services accept only `-service-name` and `-config` flags now — everything else (port, consul, DB, kafka) comes from the YAML config (and optional `DB_DSN` / `CONSUL_ADDRESS` env vars)
- Errors are wrapped with `fmt.Errorf("context: %w", err)` and checked with `errors.Is()`. Repository-layer "not found" is `repository.ErrNotFound` per service; controllers translate it to their own `ErrNotFound` at the boundary.
