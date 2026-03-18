# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SMPP Simulator - A frontend-backend separated application for testing SMS platform SMPP protocol functionality. Provides a mock SMPP server for development and testing purposes.

**Tech Stack:**
- Backend: Go + Gin + SQLite
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
├── cmd/server/main.go       # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── handler/             # HTTP handlers (Gin)
│   ├── model/               # Data models
│   ├── repository/          # SQLite data access
│   ├── service/             # Business logic
│   └── smpp/                # SMPP protocol implementation
│       ├── server.go        # TCP server
│       ├── pdu.go           # PDU encoding/decoding
│       └── session.go       # Session management
└── pkg/utils/               # Shared utilities
```

**Key Components:**
- `smpp.Server`: TCP server listening on port 2775, handles SMPP protocol
- `smpp.PDU`: Protocol Data Unit encoding/decoding for SMPP commands
- `repository.*`: SQLite-based data persistence for sessions and messages

### Frontend Structure

```
frontend/
├── src/
│   ├── api/                 # API client (Axios)
│   ├── components/          # Reusable Vue components
│   ├── views/               # Page components
│   │   ├── Dashboard.vue    # Statistics overview
│   │   ├── Sessions.vue     # SMPP connections
│   │   ├── Messages.vue     # Message list
│   │   └── Config.vue       # Mock configuration
│   ├── stores/              # Pinia state management
│   ├── router/              # Vue Router config
│   └── utils/               # WebSocket client
└── vite.config.ts           # Vite configuration with proxy
```

## API Endpoints

### Sessions
- `GET /api/sessions` - List all SMPP connections
- `DELETE /api/sessions/:id` - Disconnect a session

### Messages
- `GET /api/messages` - List messages (pagination, filters)
- `GET /api/messages/:id` - Get message details
- `POST /api/messages/:id/deliver` - Trigger delivery report

### Stats & Config
- `GET /api/stats` - Get statistics
- `GET /api/mock/config` - Get mock configuration
- `PUT /api/mock/config` - Update mock configuration

### WebSocket
- `GET /ws` - WebSocket connection for real-time updates

**Events:**
- `session_connect` - New SMPP connection
- `session_disconnect` - Connection closed
- `message_received` - New message received
- `message_delivered` - Message delivered

## SMPP Protocol

Supported PDUs:
- `bind_transmitter`, `bind_receiver`, `bind_transceiver`
- `unbind`
- `submit_sm`
- `deliver_sm` (delivery reports)
- `enquire_link`

Default port: 2775

## Environment Variables

### Backend
- `SMPP_HOST` - SMPP server host (default: 0.0.0.0)
- `SMPP_PORT` - SMPP server port (default: 2775)
- `HTTP_HOST` - HTTP API host (default: 0.0.0.0)
- `HTTP_PORT` - HTTP API port (default: 8080)
- `DB_PATH` - SQLite database path (default: ./smpp.db)

## Development Notes

- The SMPP server accepts any system_id/password for bind requests (no authentication)
- Mock configuration controls auto-response behavior and delivery reports
- SQLite uses WAL mode for better concurrent access
- Frontend dev server proxies `/api` and `/ws` to backend
