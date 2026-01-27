# BeeCount Cloud - AI Coding Instructions

## Architecture Overview

BeeCount Cloud is a **Go microservices architecture** for multi-user personal finance management with data synchronization, JWT authentication, and object storage.

### Core Components

**Microservices** (each in `services/{name}`):
- `gateway` - HTTP/REST API gateway (Gin), routes requests to backend services via gRPC
- `auth` - JWT token generation/validation and refresh token management
- `beecount` - Main service orchestrator (manages all services as child processes)
- `business` - Ledger/transaction business logic (Sync, CRUD operations)
- `config` - Distributed configuration service with file watching (Viper)
- `storage` - File upload/download/deletion with S3-compatible backends
- `log` - Centralized logging service
- `firewall` - Request filtering and access control

**Common Layer** (`common/`):
- gRPC protocol definitions in `proto/` (protobuf3 format)
- Transport abstraction layer (`transport/`) - Windows named pipes → Unix sockets → TCP fallback

### Service Communication Pattern

```
HTTP Client
    ↓
[Gateway Service] ←--gRPC-→ [Auth] [Business] [Config] [Storage] [Log] [Firewall]
    ↑                              ↑
    └──────────────────────────────┘
    (Uses platform-specific IPC)
```

**Key IPC Strategy**: Gateway uses platform detection to choose transport:
- **Windows**: Named pipes (`\\.\pipe\{service}`), fallback to TCP
- **Linux/macOS**: Unix domain sockets (`/tmp/beecount_{service}.sock`)
- **Default**: TCP localhost (port 50051-50056)

**Service Discovery**: Hardcoded addresses in each service; configured via `getSocketPath()` calls and environment variables.

---

## Project Structure & Build System

### Service Module Layout

Each service follows this pattern (e.g., `services/auth/`):
```
cmd/main.go                    # Entry point, gRPC server setup
internal/
  auth_service.go             # gRPC service implementation (handles RPC methods)
  (service-specific files)     # Business logic, helpers
pkg/
  logger/                      # Logging wrapper
  i18n/                        # Internationalization
go.mod                         # Service-specific dependencies
```

### Proto Files & Code Generation

- **Proto sources**: `common/proto/{service}/{service}.proto` (proto3 syntax)
- **Generated code**: `common/proto/{service}/{service}.pb.go` and `{service}_grpc.pb.go`
- **Generation command**:
  ```powershell
  .\scripts\generate_proto.ps1  # Windows
  bash scripts/generate_proto.sh # Linux/macOS
  ```
  Requires: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`

### Build & Execution

**Commands** (see `Makefile` for details):
- `make build` - Compiles all services to `build/bin/`, copies resources (config, web, i18n)
- `make start` - Runs BeeCount-Cloud main executable which spawns all services
- `make test` - Runs `go test ./...` across all services
- `make clean` - Removes `build/` directory

**Resources copied to `build/`**:
- `web/` - Static files (index.html, swagger.json, redoc.standalone.js)
- `config/` - YAML configuration files (database.yaml, cors.yaml, jwt.yaml, log.yaml, server.yaml, storage.yaml)
- `i18n/` - Localization files (en-US.json, zh-CN.json)

**Main Orchestrator**: `services/beecount/cmd/main.go` → CLI with Cobra commands:
- `start [--all|--background]` - Start services
- `stop` - Stop all services
- `restart` - Restart services
- `status` - Check service status
- `health` - Health check

---

## Critical Developer Workflows

### Adding a New gRPC Service

1. **Define RPC contract** in `common/proto/{service}/{service}.proto`:
   ```protobuf
   service MyService {
     rpc DoSomething(Request) returns (Response);
   }
   ```

2. **Regenerate code**: Run proto generation script

3. **Implement service** in `services/{service}/internal/{service}_service.go`:
   ```go
   type MyServiceImpl struct {
     proto.UnimplementedMyServiceServer
     common.UnimplementedHealthCheckServiceServer
   }
   func (s *MyServiceImpl) DoSomething(ctx context.Context, req *proto.Request) (*proto.Response, error) {
     // Implementation
   }
   ```

4. **Register with gRPC server** in `cmd/main.go`:
   ```go
   proto.RegisterMyServiceServer(grpcServer, myServiceImpl)
   common.RegisterHealthCheckServiceServer(grpcServer, myServiceImpl)
   ```

5. **Update gateway** (`services/gateway/internal/api_gateway.go`) to add client:
   ```go
   grpcConfig.MyServiceAddr = getSocketPath("myservice")
   g.myServiceClient = proto.NewMyServiceClient(conn)
   ```

6. **Add HTTP route** in gateway for the new RPC method

### Service Integration in Gateway

- Import proto: `"github.com/fishdivinity/BeeCount-Cloud/common/proto/{service}"`
- Create gRPC client and connection in `ConfigureGRPCClients()`
- Each HTTP endpoint calls corresponding gRPC method
- Use `g.Transport.NewDialer()` to support cross-platform IPC

### Proto Conventions

- Package: `option go_package = "github.com/fishdivinity/BeeCount-Cloud/common/proto/{service}";`
- Use `common.Response` and `common.Request` wrapper types for consistency
- Map fields for flexible key-value data (e.g., `map<string, string> tags`)
- Always import `common/common.proto` for shared types
- All services must implement `HealthCheckService` (Check, Watch RPCs)

### Configuration Management

- **Config service** loads YAML from `config/` directory via Viper
- Changes trigger file watcher events (fsnotify)
- Gateway polls config service for updates
- Environment variables override YAML via `Viper.AutomaticEnv()` with `_` separator (`DB_HOST` → `db.host`)

---

## Project-Specific Patterns & Conventions

### Error Handling
- Use gRPC status codes: `status.Errorf(codes.InvalidArgument, "message")`
- Wrap with context in middleware
- Log via `logger` package (not `fmt` or `log`)

### Logging
- Use `github.com/fishdivinity/BeeCount-Cloud/services/{service}/pkg/logger`
- Examples: `logger.Info()`, `logger.Error()`, `logger.Warning()`
- Logs written to `logs/{service}.log` and rotated automatically

### Internationalization (i18n)
- Use `github.com/fishdivinity/BeeCount-Cloud/services/beecount/pkg/i18n`
- JSON translation files: `build/i18n/en-US.json`, `build/i18n/zh-CN.json`
- Pattern: `i18n.T("key.subkey", lang)`

### Transport Abstraction
- Services use `transport.NewTransportWithFallback()` for cross-platform IPC
- Listener creation: `trans.NewListener(address)` handles platform differences
- Dial: `trans.NewDialer()` returns appropriate dialer (pipes/sockets/TCP)

### Service Lifecycle
- Services spawn as child processes from `BeeCount-Cloud` orchestrator
- PID files stored in `pids/` directory
- Logs in `logs/{service}.log`
- Automatic restart on crash (via ServiceManager)
- Graceful shutdown via context cancellation and signal handling (SIGTERM, SIGINT)

---

## External Dependencies & Integration Points

### Key Dependencies (per `go.mod`)
- `google.golang.org/grpc` - RPC framework
- `google.golang.org/protobuf` - Serialization
- `github.com/spf13/cobra` - CLI framework (beecount orchestrator)
- `github.com/spf13/viper` - Config management (config service)
- `github.com/gin-gonic/gin` - HTTP router (gateway)
- `github.com/fsnotify/fsnotify` - File watching (config service)

### Docker Deployment
- Multi-stage build: builder stage (Go 1.25.6-alpine) → runtime stage (alpine)
- All services compiled statically with `-ldflags="-s -w"` (small binary size)
- Resource copied to image during build
- No dependency on external binaries (ca-certificates, tzdata only)

### Health Checks
- All services implement `HealthCheckService.Check(ctx)` → `HealthCheckResponse`
- Called by orchestrator to detect failures
- Used in Kubernetes liveness probes

---

## Common Pitfalls & Debugging Tips

1. **Proto generation missing** - Run `generate_proto.ps1/sh` after modifying .proto files; rebuild references fail otherwise
2. **Service discovery timeout** - Check `localhost:50051-50056` are accessible; on Windows, verify named pipes (`\\.\pipe\{service}`) exist
3. **Cross-service calls hang** - Verify config service has correct addresses; use `gateway.ConfigureGRPCClients()` to test connections
4. **Config changes not propagating** - Check file watcher is active in config service; look for fsnotify errors in logs
5. **Build dir missing** - Run `make clean && make build` to ensure all resources copied
6. **Service crashes** - Check `logs/{service}.log` for errors; verify socket/pipe paths exist and are writable

---

## When to Reference Which Files

- **gRPC interfaces**: `common/proto/{service}/{service}.proto`
- **Service implementation**: `services/{service}/internal/{service}_service.go`
- **HTTP routes**: `services/gateway/internal/api_gateway.go` (SetupRoutes method)
- **CLI commands**: `services/beecount/internal/commands/*.go`
- **Build logic**: `Makefile` or `scripts/build.ps1`
- **Transport layer**: `common/transport/transport.go`, `common/transport/{windows_pipe,unix,tcp}_transport.go`
- **Config loading**: `services/config/internal/config_service.go`
