# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Alby Hub is a self-sovereign Lightning Network wallet hub implementing the NIP-47 (Nostr Wallet Connect) protocol. It allows users to connect multiple applications to their Lightning node while maintaining self-custody.

**Key Features:**

- Dual mode operation: HTTP server (web app) + Wails desktop app
- Multi-backend Lightning support: embedded LDK, LND, Phoenixd, Cashu
- NWC (Nostr Wallet Connect) protocol implementation
- Isolated app connections with budget controls
- Self-custodial transaction management

## Development Commands

### Backend Development

```bash
# HTTP server mode
go run cmd/http/main.go

# Desktop app development
wails dev -tags "wails"

# Build HTTP production binary
go build -o main cmd/http/main.go

# Build desktop production app
wails build -tags "wails"
```

### Frontend Development

```bash
cd frontend

# Install dependencies
yarn install

# Development with HTTP backend
yarn dev:http

# Development with Wails backend
yarn dev:wails

# Production builds
yarn build:http      # HTTP mode
yarn build:wails     # Desktop mode

# Linting and formatting
yarn lint           # Full lint check
yarn lint:js        # JavaScript/TypeScript only
yarn lint:js:fix    # Fix linting issues
yarn tsc:compile    # TypeScript compilation check
yarn format         # Check formatting
yarn format:fix     # Fix formatting
```

### Testing

```bash
# Run all tests (with integration tag required for this project)
go test -tags=integration ./...

# Run all tests with verbose output
go test -v -tags=integration ./...

# Run tests for specific package
go test -tags=integration ./nip47/...

# Run tests for a specific file
go test -tags=integration ./nip47/controllers/

# Run specific test
go test -tags=integration ./nip47/controllers/ -run TestHandleMultiPayInvoiceEvent

# Run specific test with verbose output
go test -v -tags=integration ./nip47/controllers/ -run TestHandleMultiPayInvoiceEvent

# Clean test cache and run all tests
go clean -testcache && go test -tags=integration ./...

# Clean test cache and run all tests with verbose output
go clean -testcache && go test -v -tags=integration ./...

# Test with PostgreSQL (requires TEST_DATABASE_URI)
export TEST_DATABASE_URI="postgresql://user:password@localhost:5432/postgres"
go test -tags=integration ./...

# Generate mocks
mockery
```

### Database Management

```bash
# Migrate from SQLite to PostgreSQL
go run cmd/db_migrate/main.go -from .data/nwc.db -to postgresql://user:pass@localhost:5432/nwc
```

## Architecture Overview

### Core Service Pattern

Alby Hub follows a **service-oriented architecture** with event-driven communication:

```
Frontend (React) ←→ HTTP/Wails API ←→ Core Service
                                           ↓
                                     Event Publisher
                                           ↓
                        ┌─────────────────────────────────┐
                        ↓         ↓         ↓             ↓
                   NIP-47    LNClient   Transactions   Database
                  Service     Layer      Service       (SQLite/PG)
```

### Key Components

**Service Layer (`/service/`):**

- Central orchestrator managing all components
- Event-driven pub/sub messaging
- Configuration and secrets management

**LN Client Abstraction (`/lnclient/`):**

- Unified interface for different Lightning implementations
- Supports LDK (embedded), LND, Phoenixd, Cashu backends

**NIP-47 Implementation (`/nip47/`):**

- Nostr Wallet Connect protocol handlers
- NIP-04 encryption for secure communication
- WebSocket relay management

**Transaction Management (`/transactions/`):**

- Internal transaction database with metadata
- Self-payment support for isolated app connections
- Budget controls per application

**HTTP API (`/http/` & `/api/`):**

- REST endpoints for frontend communication
- JWT authentication
- Business logic routing

### Key Directories

```
/cmd/                   # Application entry points
/frontend/              # React frontend
/service/              # Core service implementation
/nip47/               # NWC protocol handlers
/lnclient/            # Lightning client abstraction
/transactions/        # Transaction management
/db/                  # Database models & migrations
/http/                # HTTP service layer
/events/              # Event publishing system
/config/              # Configuration management
/alby/                # Alby OAuth integration
```

## Development Patterns

### Interface-Based Design

- All major components use interfaces for testability
- LNClient interface allows swappable Lightning implementations
- Service interfaces enable clean dependency injection

### Event-Driven Architecture

- Central event publisher coordinates component communication
- Events flow: LNClient → Transactions → NIP-47 → Alby OAuth
- Key events: payments, channel updates, app connections

### Database Strategy

- GORM for ORM with SQLite/PostgreSQL support
- Migration system with versioning
- Transaction metadata storage separate from LN backend

## Configuration

**Environment Variables:**

- `DATABASE_URI`: SQLite file or PostgreSQL connection string
- `WORK_DIR`: Data directory (default: $XDG_DATA_HOME/albyhub)
- `PORT`: HTTP server port (default: 8080)
- `LOG_LEVEL`: Logging verbosity (default: 4)
- `RELAY`: Nostr relay URL
- `NETWORK`: Bitcoin network (mainnet/testnet/signet/regtest)

**LDK-Specific:**

- `LDK_ESPLORA_SERVER`: Esplora server URL
- `LDK_LISTENING_ADDRESSES`: P2P listening addresses
- `LDK_BITCOIND_RPC_*`: Bitcoin RPC connection

## Testing Strategy

- Unit tests with testify/mock framework
- Mock generation via vektra/mockery
- PostgreSQL integration tests with pgtestdb
- Docker compose for test database setup

## Security Considerations

- AES encryption for sensitive data (seed phrases)
- JWT authentication for HTTP API
- NIP-04 encryption for Nostr communication
- Never log sensitive data or private keys

## Build Targets

- **HTTP Mode**: Standard web application
- **Desktop Mode**: Wails v2 native app
- **Docker**: Container deployment
- **Multiple architectures**: Linux (x86_64, arm64), macOS, Windows
