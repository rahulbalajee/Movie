# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build all services
go build ./metadata/cmd ./rating/cmd ./movie/cmd

# Run all tests
go test ./...

# Run benchmarks (serialization size comparison)
go test ./cmd/sizecompare/... -bench=.

# Run a single test by name
go test ./path/to/package -run TestName

# Generate protobuf code (requires protoc, protoc-gen-go, protoc-gen-go-grpc)
protoc -I api/ api/movie.proto --go_out=gen/ --go_opt=paths=source_relative --go-grpc_out=gen/ --go-grpc_opt=paths=source_relative
```

## Architecture

Three microservices communicating via gRPC and HTTP, with Consul-based service discovery:

- **metadata** (gRPC, port 8081) — stores movie metadata (title, description, director)
- **rating** (gRPC, port 8082) — stores and aggregates user ratings
- **movie** (gRPC, port 8083) — aggregation service that calls metadata + rating to build complete movie details

All services register with Consul on startup, report health via TTL checks every 1s, and deregister on graceful shutdown (SIGINT/SIGTERM).

### Internal structure per service

Each service follows the same layered layout: `cmd/main.go` → `internal/handler/` → `internal/controller/` → `internal/repository/`. The movie service replaces repository with `internal/gateway/` since it fetches data from the other two services rather than storing its own.

### Shared packages

`pkg/discovery/` defines the `Registry` interface (Register, Deregister, ServiceAddresses, ReportHealthyState) with two implementations: `consul/` (production, requires Consul at localhost:8500) and `memory/` (in-process, for testing).

### Protobuf / gRPC

Proto definitions live in `api/movie.proto`. Generated Go code lives in `gen/` (`movie.pb.go`, `movie_grpc.pb.go`). All three services serve gRPC. Legacy HTTP handlers and gateways still exist under `internal/handler/http/` and `movie/internal/gateway/.../http/` but are no longer wired into `cmd/main.go`.

### Service communication flow

```
movie service --gRPC--> metadata service
movie service --gRPC--> rating service
```

## Key conventions

- Module path: `github.com/rahulbalajee/Movie`
- All repositories are currently in-memory (map + `sync.RWMutex`)
- Controllers accept interfaces, not concrete types (repository pattern + dependency injection)
- Services accept `-port`, `-service-name`, and `-consul-addr` flags
- Errors are wrapped with `fmt.Errorf("context: %w", err)` and checked with `errors.Is()`
