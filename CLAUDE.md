# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SMPP Simulator - A frontend-backend separated application for testing SMS platform SMPP protocol functionality. Provides a mock SMPP server for development and testing purposes.

**Tech Stack:**
- Backend: Go + Gin + SQLite + JWT
- Frontend: Vue 3 + TypeScript + Element Plus + Pinia
- Build: Vite, Docker

## Common Commands

### Backend (Go)

```bash
cd backend

# Install dependencies
go mod download

# Run development server
go run cmd/server/main.go

# Build
go build -o smpp-simulator ./cmd/server

# Run tests
go test ./...
```

### Frontend (Vue)

```bash
cd frontend

# Install dependencies
pnpm install

# Run development server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview
```

### Docker

```bash
# Build and run all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild
docker-compose up -d --build
```

## Architecture

### Backend Structure

```
backend/
├── cmd/server/main.go          # Application entry point
├── config.example.yaml         # Configuration template
├── internal/
│   ├── config/                 # Configuration management (YAML + env)
│   ├── handler/                # HTTP handlers (Gin)
│   │   ├── auth.go             # Login/status endpoints
│   │   ├── message.go          # Message CRUD + stats
│   │   ├── session.go          # Session management
│   │   └── websocket.go        # WebSocket hub
│   ├── middleware/             # HTTP middleware
│   │   └── auth.go             # JWT authentication
│   ├── model/                  # Data models
│   ├── repository/             # SQLite data access
│   └── smpp/                   # SMPP protocol implementation
│       ├── server.go           # TCP server
│       ├── pdu.go              # PDU encoding/decoding
│       └── session.go          # Session state
└── pkg/
    └── jwt/jwt.go              # JWT utilities
```

**Key Components:**
- `smpp.Server`: TCP server listening on port 2775, handles SMPP protocol
- `smpp.PDU`: Protocol Data Unit encoding/decoding for SMPP commands
- `repository.*`: SQLite-based data persistence for sessions and messages
- `middleware.AuthMiddleware`: JWT authentication for protected routes

### Frontend Structure

```
frontend/
├── src/
│   ├── api/index.ts            # API client (Axios) with auth interceptors
│   ├── views/                  # Page components
│   │   ├── Dashboard.vue       # Statistics overview (public)
│   │   ├── Sessions.vue        # SMPP connections (protected)
│   │   ├── Messages.vue        # Message list with filters (public)
│   │   ├── Config.vue          # Mock configuration (protected)
│   │   └── Login.vue           # Login page
│   ├── stores/                 # Pinia state management
│   │   ├── index.ts            # Session, message, stats, config stores
│   │   └── auth.ts             # Authentication store
│   ├── router/index.ts         # Vue Router with auth guards
│   └── utils/websocket.ts      # WebSocket client with heartbeat
└── vite.config.ts              # Vite configuration with proxy
```

## Configuration

### Configuration Methods (Priority: env > config file > defaults)

**1. Config File (config.yaml):**
```yaml
smpp_host: "0.0.0.0"
smpp_port: "2775"
http_host: "0.0.0.0"
http_port: "8080"
db_path: "./smpp.db"
admin_password: "admin123"
jwt_secret: "your-secret-key"
jwt_expiry: 24
```

**2. Environment Variables:**
| Variable | Default | Description |
|----------|---------|-------------|
| `SMPP_HOST` | 0.0.0.0 | SMPP server host |
| `SMPP_PORT` | 2775 | SMPP server port |
| `HTTP_HOST` | 0.0.0.0 | HTTP API host |
| `HTTP_PORT` | 8080 | HTTP API port |
| `DB_PATH` | ./smpp.db | SQLite database path |
| `ADMIN_PASSWORD` | admin123 | Admin login password |
| `JWT_SECRET` | smpp-simulator-secret-key | JWT signing key |
| `JWT_EXPIRY` | 24 | Token expiry (hours) |
| `CONFIG_PATH` | auto | Custom config file path |

## API Endpoints

### Public Endpoints (No Auth Required)

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/auth/login | Login, returns JWT token |
| GET | /api/auth/status | Check token validity |
| GET | /api/stats | Get statistics |
| GET | /api/messages | List messages (with filters) |
| GET | /api/messages/:id | Get message details |

### Protected Endpoints (Auth Required)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/sessions | List SMPP connections |
| DELETE | /api/sessions/:id | Disconnect a session |
| POST | /api/messages/:id/deliver | Mark message as delivered |
| POST | /api/messages/:id/fail | Mark message as failed |
| GET | /api/mock/config | Get mock configuration |
| PUT | /api/mock/config | Update mock configuration |

### Message Filters

Query parameters for `GET /api/messages`:
- `status` - Filter by status (pending/delivered/failed)
- `source_addr` - Filter by sender (partial match)
- `dest_addr` - Filter by receiver (partial match)
- `start_time` - Filter from datetime (YYYY-MM-DD HH:mm:ss)
- `end_time` - Filter to datetime
- `page` - Page number
- `page_size` - Items per page (max 100)

### WebSocket

- **Endpoint:** `GET /ws`
- **Heartbeat:** Client sends `{"type":"ping"}` every 30s, server responds with `{"type":"pong"}`

**Events:**
- `session_connect` - New SMPP connection
- `session_disconnect` - Connection closed
- `message_received` - New message received
- `message_delivered` - Message delivered

## Authentication

### User Permissions

| Page | Unauthenticated | Authenticated |
|------|:---------------:|:-------------:|
| Dashboard | ✅ | ✅ |
| Messages | ✅ | ✅ |
| Sessions | ❌ | ✅ |
| Config | ❌ | ✅ |

### Default Credentials
- Username: `admin`
- Password: `admin123` (change via config)

## SMPP Protocol

### Supported PDUs
- `bind_transmitter`, `bind_receiver`, `bind_transceiver`
- `unbind`
- `submit_sm`
- `deliver_sm` (delivery reports)
- `enquire_link`

### Message Encoding
| data_coding | Encoding | Support |
|-------------|----------|:-------:|
| 0 | GSM7/ASCII | ✅ |
| 8 | UCS2 (UTF-16BE) | ✅ |

Default port: 2775

## Development Notes

- SMPP server accepts any system_id/password for bind requests
- Mock configuration controls auto-response behavior and delivery reports
- SQLite uses WAL mode for better concurrent access
- Frontend dev server proxies `/api` and `/ws` to backend
- WebSocket uses heartbeat (30s) to prevent connection timeout
- Auto delivery report skips if message status was manually changed
