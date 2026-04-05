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

### E2E Testing (端到端测试)

**重要：开发完成后，AI 应自动运行这些测试来验证功能。**

```bash
# API 端到端测试 (HTTP API)
./scripts/test-e2e.sh

# SMPP 完整测试 (协议 + API)
./scripts/test-smpp.sh

# SMPP 客户端工具 (手动测试)
python3 scripts/smpp_client.py --host 127.0.0.1 --port 2775 test
```

**测试覆盖：**
- `test-e2e.sh`: 认证、Session、消息、模板、Mock 配置、系统配置等 API
- `test-smpp.sh`: SMPP 绑定、发送消息、消息存储验证、状态更新

**开发流程建议：**
1. 编写代码
2. 运行 `./scripts/test-e2e.sh` 验证 API
3. 运行 `./scripts/test-smpp.sh` 验证 SMPP 协议
4. 根据错误日志修复问题
5. 重复直到所有测试通过

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
│   │   ├── mock.go             # Mock configuration
│   │   ├── data.go             # Data management (clear data)
│   │   ├── send_message.go     # Admin send message
│   │   ├── outbound.go         # Outbound SMPP connection handler
│   │   └── websocket.go        # WebSocket hub
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go             # JWT authentication
│   │   └── ratelimit.go        # Rate limiting for login
│   ├── model/                  # Data models
│   ├── repository/             # SQLite data access
│   └── smpp/                   # SMPP protocol implementation
│       ├── server.go           # TCP server (passive mode)
│       ├── client.go           # SMPP client (active mode)
│       ├── pdu.go              # PDU encoding/decoding
│       └── session.go          # Session state
└── pkg/
    └── jwt/jwt.go              # JWT utilities
```

**Key Components:**
- `smpp.Server`: TCP server listening on port 2775, handles incoming SMPP connections
- `smpp.Client`: SMPP client for initiating outbound connections to remote SMSCs
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
cors_origins: "*"
login_rate_limit: 5
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
| `CORS_ORIGINS` | * | Allowed CORS origins (comma-separated) |
| `LOGIN_RATE_LIMIT` | 5 | Max login attempts per minute per IP |
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
| DELETE | /api/data/messages | Delete all messages |
| DELETE | /api/data/sessions | Delete all sessions |
| DELETE | /api/data/all | Delete all messages and sessions |
| GET | /api/send/receivers | List sessions that can receive messages |
| POST | /api/send | Send message to a connected session |
| GET | /api/outbound | List outbound SMPP connections |
| POST | /api/outbound/connect | Connect to remote SMSC |
| DELETE | /api/outbound/:id | Disconnect outbound session |
| POST | /api/outbound/:id/send | Send message via outbound session |

### Health Check

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check endpoint |

### Message Filters

Query parameters for `GET /api/messages`:
- `status` - Filter by status (pending/delivered/failed)
- `source_addr` - Filter by sender (partial match)
- `dest_addr` - Filter by receiver (partial match)
- `start_time` - Filter from datetime (YYYY-MM-DD HH:mm:ss)
- `end_time` - Filter to datetime
- `page` - Page number
- `page_size` - Items per page (max 100)

### Send Message

Request body for `POST /api/send`:
```json
{
  "session_id": "session-uuid",
  "source_addr": "10086",
  "dest_addr": "13800138000",
  "content": "Message content",
  "encoding": "GSM7" // or "UCS2"
}
```

### Outbound SMPP Connections

Connect to a remote SMSC:
```json
// POST /api/outbound/connect
{
  "host": "192.168.1.100",
  "port": "2775",
  "system_id": "username",
  "password": "password",
  "bind_type": "transceiver" // transmitter, receiver, or transceiver
}
```

Send message via outbound connection:
```json
// POST /api/outbound/:id/send
{
  "source_addr": "10086",
  "dest_addr": "13800138000",
  "content": "Message content",
  "encoding": "GSM7" // or "UCS2"
}
```

### WebSocket

- **Endpoint:** `GET /ws`
- **Authentication:** Optional (JWT token via query parameter `?token=<jwt>` or `Authorization` header)
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

## Security Features

- **Login Rate Limiting**: Max 5 attempts per minute per IP by default
- **CORS Configuration**: Configurable via `CORS_ORIGINS` env var
- **WebSocket Origin Validation**: Checks origin against allowed list
- **Secure ID Generation**: Uses `crypto/rand` for session/message IDs
- **Docker Security**: Container runs as non-root user with health checks

## API Documentation

OpenAPI 3.0 specification available at `docs/openapi.yaml`

View with:
```bash
npx redocly preview-docs docs/openapi.yaml
```
