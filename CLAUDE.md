# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build all services + tools
go build ./metadata/cmd ./rating/cmd ./movie/cmd ./cmd/ratingproducer ./cmd/sizecompare

# Run all tests
go test ./...

# Run benchmarks (serialization size comparison)
go test ./cmd/sizecompare/... -bench=.

# Run a single test by name
go test ./path/to/package -run TestName

# Generate protobuf code (requires protoc, protoc-gen-go, protoc-gen-go-grpc)
protoc -I api/ api/movie.proto --go_out=gen/ --go_opt=paths=source_relative --go-grpc_out=gen/ --go-grpc_opt=paths=source_relative

# Bring up Kafka + Zookeeper for the rating ingester (from cmd/ratingproducer)
docker-compose -f cmd/ratingproducer/docker-compose.yaml up -d

# Apply MySQL schema (DSN hardcoded as root:password@/movieexample in cmd/main.go)
mysql -u root -p movieexample < schema/schema.sql

# Seed Kafka with rating events from cmd/ratingproducer/ratingsdata.json
go run ./cmd/ratingproducer
```

## Architecture

Three microservices communicating via gRPC and HTTP, with Consul-based service discovery:

- **metadata** (gRPC, port 8081) — stores movie metadata (title, description, director)
- **rating** (gRPC, port 8082) — stores and aggregates user ratings
- **movie** (gRPC, port 8083) — aggregation service that calls metadata + rating to build complete movie details

All services register with Consul on startup, report health via TTL checks every 1s, and deregister on graceful shutdown (SIGINT/SIGTERM).

### Internal structure per service

Each service follows the same layered layout: `cmd/main.go` → `internal/handler/` → `internal/controller/` → `internal/repository/`. The movie service replaces repository with `internal/gateway/` since it fetches data from the other two services rather than storing its own.

### Persistence

Both metadata and rating now persist to MySQL (`internal/repository/mysql/`) — the in-memory repos under `internal/repository/memory/` are still present for tests. Schema lives in `schema/schema.sql` (single `movieexample` database with `movies` and `ratings` tables). The MySQL DSN is currently hardcoded in each service's `cmd/main.go` (`root:password@/movieexample`) — change there, not via env.

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

## Key conventions

- Module path: `github.com/rahulbalajee/Movie`
- Production repos are MySQL; the `memory` repos remain for tests. The cache layer reuses the same repo interface, which is why the metadata controller can accept both.
- Controllers accept interfaces, not concrete types (repository pattern + dependency injection)
- Services accept `-port`, `-service-name`, and `-consul-addr` flags. The rating service additionally accepts `-kafka-addr` (default `localhost:9092`).
- Errors are wrapped with `fmt.Errorf("context: %w", err)` and checked with `errors.Is()`. Repository-layer "not found" is `repository.ErrNotFound` per service; controllers translate it to their own `ErrNotFound` at the boundary.
