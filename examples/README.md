# AssetTrack

Production-ready Asset Management application built with Go.

## Features

- **REST API** - Full CRUD for assets and maintenance records
- **Web UI** - Dynamic server-rendered UI using minty/mintydyn
- **Dark Mode** - Toggle with localStorage persistence
- **Chi Router** - Clean routing with middleware support
- **Structured Logging** - slog with JSON or text output
- **Graceful Shutdown** - Clean handling of SIGINT/SIGTERM

## Quick Start

```bash
go run ./cmd/assettrack

# Browse UI: http://localhost:31271/
# API:       http://localhost:31271/api/
```

## Project Structure

```
assettrack/
├── cmd/
│   └── assettrack/
│       └── main.go          # Entry point, server setup
├── internal/
│   ├── api/
│   │   └── handlers.go      # REST API handlers
│   ├── ui/
│   │   ├── handlers.go      # Web UI page handlers
│   │   └── components.go    # UI components (minty)
│   ├── middleware/
│   │   └── middleware.go    # HTTP middleware
│   ├── store/
│   │   └── memory.go        # Data storage (in-memory)
│   └── models/
│       └── models.go        # Domain types
├── minty/                   # minty library (embedded)
├── go.mod
└── README.md
```

## Configuration

| Env Variable | Flag | Default | Description |
|--------------|------|---------|-------------|
| `PORT` | `-port` | 31271 | HTTP port |
| `LOG_LEVEL` | `-log-level` | info | debug, info, warn, error |
| `LOG_FORMAT` | `-log-format` | text | text, json |

```bash
# Examples
PORT=8080 go run ./cmd/assettrack
go run ./cmd/assettrack -port 8080 -log-level debug -log-format json
```

## API Endpoints

### Assets

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/assets` | List assets (with filters) |
| GET | `/api/assets/{id}` | Get single asset |
| POST | `/api/assets` | Create asset |
| PUT | `/api/assets/{id}` | Update asset |
| DELETE | `/api/assets/{id}` | Delete asset |
| GET | `/api/assets/stats` | Get aggregate stats |
| GET | `/api/assets/{id}/maintenance` | Get maintenance records |

### Maintenance

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/maintenance` | List all records |
| POST | `/api/maintenance` | Create record |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |

### Query Parameters

```bash
# Filter assets
GET /api/assets?status=active
GET /api/assets?category=Laptops
GET /api/assets?department=Engineering
GET /api/assets?search=macbook
```

### Example Responses

```json
// GET /api/assets
{
  "data": [...],
  "meta": { "total": 10 }
}

// GET /api/assets/A001
{
  "data": {
    "id": "A001",
    "name": "MacBook Pro 16\"",
    ...
  }
}

// Error response
{
  "error": "Asset not found"
}
```

## UI Pages

| Path | Description |
|------|-------------|
| `/` | Dashboard with stats |
| `/assets` | Asset list with filtering |
| `/assets/{id}` | Asset detail with tabbed form |
| `/maintenance` | Maintenance records |
| `/reports` | Report cards |
| `/settings` | Settings with tabs |

## Building for Production

```bash
# Build binary
go build -o assettrack ./cmd/assettrack

# With version info
go build -ldflags "-X main.version=1.0.0" -o assettrack ./cmd/assettrack

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o assettrack-linux ./cmd/assettrack
```

## Deployment

### Systemd Service

```ini
# /etc/systemd/system/assettrack.service
[Unit]
Description=AssetTrack
After=network.target

[Service]
Type=simple
User=assettrack
Environment=PORT=31271
Environment=LOG_FORMAT=json
ExecStart=/opt/assettrack/assettrack
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Caddy Reverse Proxy

```caddyfile
assettrack.example.com {
    reverse_proxy localhost:31271
}
```

### Docker

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o assettrack ./cmd/assettrack

FROM alpine:3.19
COPY --from=builder /app/assettrack /usr/local/bin/
EXPOSE 31271
CMD ["assettrack"]
```

## Development

```bash
# Run with hot reload (using air)
air

# Run tests
go test ./...

# Lint
golangci-lint run
```

## Architecture Notes

### Why Chi?

- 100% compatible with `net/http`
- Clean route groups and middleware chaining
- Context-based parameter passing
- Active maintenance, stable API

### Separation of Concerns

- **api/** - Stateless REST handlers, JSON responses
- **ui/** - Server-rendered HTML with minty
- **store/** - Data layer interface (swap for database)
- **middleware/** - Reusable HTTP middleware

### Future Improvements

1. **Database** - Replace MemoryStore with PostgreSQL/SQLite
2. **Authentication** - Add JWT or session-based auth
3. **Caching** - Add Redis for session/query caching
4. **Metrics** - Add Prometheus metrics endpoint
5. **API Versioning** - `/api/v1/` prefix

## License

MIT
