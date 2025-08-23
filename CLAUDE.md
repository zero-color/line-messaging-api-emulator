# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build and Run
```bash
# Start the PostgreSQL database and run the server
make run

# Or run directly with Go
go run ./cmd/server

# Run with custom port
go run ./cmd/server --port 8080
```

### Testing
```bash
# Run all tests with race detection
make test

# Run a single test
go test -v -run TestName ./server/
```

### Code Generation
```bash
# Generate all code (OpenAPI server stubs, database code)
make generate

# Or directly with go generate
go generate ./...
```

### Database Management
```bash
# Start the PostgreSQL database
docker compose up -d

# Apply database migrations and regenerate SQLC code
make db-migrate

# Or manually
cd db && atlas schema apply --env local && sqlc generate
```

### Code Quality
```bash
# Run linters
make lint

# Or directly with golangci-lint
go tool golangci-lint run ./...
```

## Architecture Overview

### Project Structure
This is a LINE Messaging API emulator built in Go with a layered architecture:

1. **API Layer** (`api/`)
   - `adminapi/`: Admin API for bot management (no auth required)
   - `messagingapi/`: LINE Messaging API implementation (requires auth)
   - Both use OpenAPI code generation via oapi-codegen

2. **Server Layer** (`server/`)
   - Implements both `adminapi.StrictServerInterface` and `messagingapi.StrictServerInterface`
   - Each endpoint category has its own file (e.g., `bot.go`, `message.go`, `webhook.go`)
   - Central `Server` interface defined in `server.go`

3. **Database Layer** (`db/`)
   - Uses SQLC for type-safe SQL queries
   - Schema managed by Atlas migrations
   - PostgreSQL as the database
   - Queries defined in `db/queries/` directory

4. **Authentication** (`internal/auth/`)
   - Middleware for channel access token validation
   - Applied only to Messaging API routes

### Key Design Patterns

**OpenAPI-First Development**: The API is defined by OpenAPI specs in `line-openapi/` directory. Server stubs are generated using oapi-codegen with strict handlers pattern.

**Database Code Generation**: All database interactions use SQLC-generated code. SQL queries are written in `db/queries/*.sql` files and generate type-safe Go code.

**Dependency Injection**: The server accepts a `db.Querier` interface, making it testable and decoupled from the database implementation.

**Middleware Pattern**: Authentication is implemented as Chi middleware, applied selectively to protected routes.

## Code Style Guidelines
In general, you shouldn't write obvious comments like "Log the error" above error handling code, as it is redundant.
Instead, focus on writing clear and self-explanatory code that adheres to the project's conventions.

#### Go
- Follow standard Go conventions
- Write comment for public functions and types
- Each function length should be under 150 lines
- Keep readability high with proper function separation using helper private functions

### Development Workflow

1. **API Changes**: Modify OpenAPI specs → Run `make generate` → Implement new methods in `server/` package
2. **Database Changes**: Update `db/schema.sql` → Run `make db-migrate` → Write queries in `db/queries/` → Generate code
3. **Testing**: Each server component has its own test file (e.g., `bot_test.go`, `admin_test.go`)

### Environment Configuration

The server accepts configuration through command-line flags and environment variables:
- `--port`: HTTP port (default: 9090)
- `--database-url` or `DATABASE_URL`: PostgreSQL connection string

Default database connection: `postgres://postgres:password@localhost:5432/line_emulator?sslmode=disable`