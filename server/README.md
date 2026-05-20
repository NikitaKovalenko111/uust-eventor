# Eventor Server

Go backend for Eventor mobile application with JWT authentication, PostgreSQL database, and Fiber web framework.

## Architecture

```
internal/
├── auth/                    # Authentication module
│   ├── domain/             # Core models and errors
│   ├── services/           # Business logic
│   ├── storage/            # Database repositories
│   └── transport/http/     # HTTP handlers and middleware
├── platform/               # Shared infrastructure
│   ├── config/            # Configuration management
│   ├── storage/           # Database connection
│   └── types/             # Common types
└── events/                # Events management module (WIP)

cmd/
├── main.go               # Application entry point
└── migrator/             # Database migration tool
```

## Prerequisites

- Go 1.25.4 or higher
- PostgreSQL 12 or higher
- Environment variables configured in `.env`

## Setup

### 1. Environment Configuration

Create `.env` file in server root:

```env
CONFIG_PATH=./config/local.yml
```

Create `config/local.yml`:

```yaml
env: local

storage:
    db_host: "localhost"
    db_port: 5432
    db_user: "eventor"
    db_pass: "your_password"
    db_name: "eventor_db"

http_server:
    http_address: "localhost:8080"
    timeout: "4s"
    idle_timeout: "60s"

jwt:
    secret: "your-super-secret-jwt-key-min-32-chars"
    access_token_ttl: "15m"
    refresh_token_ttl: "7d"
```

### 2. Database Setup

Create PostgreSQL database:

```sql
CREATE DATABASE eventor_db;
CREATE USER eventor WITH PASSWORD 'your_password';
ALTER ROLE eventor WITH CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE eventor_db TO eventor;
```

### 3. Run Migrations

```bash
cd server
go run ./cmd/migrator --migrations-path ./internal/platform/storage/migrations
```

### 4. Build and Run Server

Development mode:

```bash
cd server
go run ./cmd/main.go
```

Production build:

```bash
cd server
go build -o server.exe ./cmd
./server.exe
```

## API Documentation

See [AUTH_API.md](AUTH_API.md) for complete authentication API documentation.

### Quick API Examples

#### Register

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "role": "user"
  }'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Refresh Token

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

## Project Structure

### internal/auth/ - Authentication Module

**Domain Layer** (`domain/`)

- `errors/errors.go`: Sentinel error definitions for auth operations
- `models/user.go`: User entity with hashed passwords
- `models/token.go`: JWT claims and token pair structures

**Service Layer** (`services/`)

- `jwt.go`: Token generation and verification with HMAC-SHA256
- `usecase/user/user.go`: User registration and login business logic
- `usecase/token/token.go`: Token lifecycle management

**Storage Layer** (`storage/repositories/`)

- `user/user.go`: User CRUD with bcrypt password hashing
- `token/token.go`: Refresh token storage with SHA256 hashing

**Transport Layer** (`transport/http/`)

- `controllers/auth/auth.go`: HTTP handlers (Login, Register, RefreshToken)
- `middlewares/jwt.go`: JWT verification middleware for protected routes

### internal/platform/ - Infrastructure

- `config/config.go`: Configuration loading from YAML and env
- `storage/storage.go`: PostgreSQL connection pooling
- `storage/migrations/`: SQL migration files
- `types/types.go`: Common type definitions

## Database Schema

### users table

```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  role user_role DEFAULT 'user',
  about TEXT,
  faculty VARCHAR,
  course INT,
  avatar_image_id VARCHAR,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);
```

### auth_tokens table

```sql
CREATE TABLE auth_tokens (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ
);
```

## Dependencies

```
github.com/gofiber/fiber/v2      - Web framework
github.com/golang-jwt/jwt/v5     - JWT tokens
github.com/lib/pq                - PostgreSQL driver
golang.org/x/crypto              - Password hashing (bcrypt)
github.com/ilyakaznacheev/cleanenv - Config loading
```

Install: `go mod download && go mod tidy`

## Features

✅ JWT authentication with AccessToken + RefreshToken
✅ Password hashing with bcrypt
✅ PostgreSQL database with migrations
✅ Structured logging with slog
✅ Clean architecture with clear layer separation
✅ Fiber web framework with middleware
✅ Token refresh mechanism
✅ User registration and login
✅ Protected routes with JWT middleware

## Error Handling

All errors are defined as domain errors and propagated using `errors.Is()`:

- `ErrInvalidCredentials`: Login failed
- `ErrEmailAlreadyExists`: Duplicate email during registration
- `ErrTokenExpired`: Token TTL exceeded
- `ErrInvalidToken`: Malformed or unsigned token
- `ErrRefreshTokenExpired`: Refresh token expired

## Logging

Operations logged with structured logging:

```
INFO auth-service: user registered email=user@example.com role=user
INFO auth-service: user logged in email=user@example.com
WARN jwt-middleware: token verification failed error=invalid_signature
ERROR auth-service: database error err=connection_refused
```

## Testing

To test auth endpoints locally:

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123","role":"user"}'

# Login (save tokens)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}' | jq

# Refresh (use response tokens)
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"{refresh_token}"}' | jq
```

## Development

### Code Style

- Follow Go idioms and conventions
- Use structured logging for all operations
- Error handling via domain errors and `errors.Is()`
- Comments for exported functions and complex logic

### Adding Features

1. Define domain models in `internal/auth/domain/models/`
2. Add business logic in `internal/auth/services/usecase/`
3. Implement repository methods in `internal/auth/storage/repositories/`
4. Create HTTP handlers in `internal/auth/transport/http/controllers/`
5. Register routes in `cmd/main.go`

## Troubleshooting

**Database connection refused**

- Check PostgreSQL is running
- Verify connection string in config
- Ensure database and user created

**Config file not found**

- Set `CONFIG_PATH` environment variable
- Verify config file exists at specified path

**Token verification fails**

- Check JWT secret matches in config
- Verify token hasn't expired
- Ensure Bearer format in Authorization header

## Future Enhancements

- [ ] Token revocation/blacklisting
- [ ] Email verification
- [ ] Password reset flow
- [ ] OAuth2 integration
- [ ] Rate limiting
- [ ] 2FA support
- [ ] Events module completion
- [ ] API documentation (Swagger)
- [ ] Automated tests
- [ ] Performance monitoring
