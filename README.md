# Chirpy

Chirpy is a Go backend for a small social-media-style application. It provides a REST-style HTTP API for user accounts, authentication, chirps, refresh-token management, and Chirpy Red membership upgrades through a webhook.

The project is a practical backend project focused on building an HTTP server with Go's `net/http`, PostgreSQL persistence, SQL migrations, type-safe database access with sqlc, password hashing, JWT authentication, refresh tokens, middleware, and webhook authentication.

## Features

- User registration and account updates
- Argon2id password hashing and verification
- JWT access-token authentication
- Refresh-token creation, validation, expiration, and revocation
- Create, list, filter, retrieve, and delete chirps
- Chirp length validation (140 characters maximum)
- Profanity filtering for chirp content
- Author-based chirp filtering with `author_id`
- Ascending/descending chirp sorting with the `sort` query parameter
- Chirpy Red membership status
- Authenticated Polka webhook for upgrading users to Chirpy Red
- API-key authentication for the Polka webhook
- PostgreSQL persistence
- SQL queries generated into type-safe Go code with sqlc
- Goose-compatible database migrations
- Static file serving through `/app/`
- File-server visit metrics
- Development-only database reset endpoint

## Prerequisites

You will need:

- Go 1.27 or later
- PostgreSQL
- Goose for running database migrations
- sqlc if you modify SQL queries or schema and need to regenerate the database code

The project currently uses Go 1.27.0 and PostgreSQL. Its main Go dependencies include JWT, Argon2id, UUID, `godotenv`, and the PostgreSQL driver.

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/Paa-Kwasi-04/chirpy.git
cd chirpy
```

### 2. Install Go dependencies

```bash
go mod download
```

### 3. Create the PostgreSQL database

Create a PostgreSQL database for Chirpy. For example:

```sql
CREATE DATABASE chirpy;
```

A typical connection string is:

```text
postgres://username:password@localhost:5432/chirpy?sslmode=disable
```

### 4. Configure environment variables

Chirpy loads configuration from a `.env` file using `godotenv`.

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
TOKEN_SECRET=replace-with-a-random-secret
POLKA_KEY=replace-with-your-polka-api-key
```

Configuration:

- `DB_URL` — PostgreSQL connection string.
- `PLATFORM` — controls the application platform/environment. The reset endpoint only works when this is `dev`.
- `TOKEN_SECRET` — secret used to sign and validate JWT access tokens.
- `POLKA_KEY` — API key used to authenticate requests to the Polka webhook.

**Do not commit `.env` or real secrets to Git.**

### 5. Run the database migrations

The migrations are stored in `internal/sql/schema/` and use Goose annotations.

Install Goose if needed:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Then run:

```bash
goose -dir internal/sql/schema postgres "$DB_URL" up
```

On Windows PowerShell, you can provide the connection string directly:

```powershell
goose -dir internal/sql/schema postgres "postgres://username:password@localhost:5432/chirpy?sslmode=disable" up
```

## Running Chirpy

Start the server with:

```bash
go run .
```

The server listens on:

```text
http://localhost:8080
```

To build an executable:

```bash
go build -o chirpy .
```

## API Endpoints

### Health Check

```http
GET /api/healthz
```

Returns a successful `204 No Content` response when the server is running.

### Create a User

```http
POST /api/users
Content-Type: application/json
```

Example body:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

The password is hashed with Argon2id before being stored.

### Update a User

```http
PUT /api/users
Authorization: Bearer <access-token>
Content-Type: application/json
```

Updates the authenticated user's email and password.

### Login

```http
POST /api/login
Content-Type: application/json
```

Example body:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

A successful login returns the user information, a JWT access token, and a refresh token.

### Refresh an Access Token

```http
POST /api/refresh
Authorization: Bearer <refresh-token>
```

Returns a new JWT access token when the refresh token exists, has not expired, and has not been revoked.

### Revoke a Refresh Token

```http
POST /api/revoke
Authorization: Bearer <refresh-token>
```

Marks the supplied refresh token as revoked.

### Create a Chirp

```http
POST /api/chirps
Authorization: Bearer <access-token>
Content-Type: application/json
```

Example body:

```json
{
  "body": "Hello from Chirpy!"
}
```

The authenticated user's ID is taken from the JWT. Chirps are limited to 140 characters, and configured profane words are replaced before the chirp is stored.

### Get Chirps

```http
GET /api/chirps
```

Returns chirps from the database.

Optional query parameters:

```text
sort=asc
sort=desc
author_id=<user-uuid>
```

Examples:

```text
GET /api/chirps?sort=desc
GET /api/chirps?author_id=<user-uuid>
GET /api/chirps?author_id=<user-uuid>&sort=desc
```

Chirps are returned in ascending creation order by default. Supplying `sort=desc` reverses the order. Supplying `author_id` limits the results to chirps belonging to that user.

### Get a Single Chirp

```http
GET /api/chirps/{chirpID}
```

Replace `{chirpID}` with the chirp's UUID.

### Delete a Chirp

```http
DELETE /api/chirps/{chirpID}
Authorization: Bearer <access-token>
```

The request must be authenticated, and only the user who owns the chirp can delete it.

## Polka Webhook

Chirpy supports a webhook for upgrading users to **Chirpy Red** membership.

```http
POST /api/polka/webhooks
Authorization: ApiKey <polka-api-key>
Content-Type: application/json
```

Expected request body:

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "<user-uuid>"
  }
}
```

The webhook:

1. Validates the `ApiKey` authorization header against `POLKA_KEY`.
2. Checks the webhook event type.
3. Parses the supplied user UUID.
4. Updates the user's `is_chirpy_red` field to `true`.

Events other than `user.upgraded` are acknowledged without changing the user's membership status.

## Admin Endpoints

### Metrics

```http
GET /admin/metrics
```

Returns an HTML page showing how many times the `/app/` file server has been visited.

### Reset

```http
POST /admin/reset
```

Deletes all users. Because chirps and refresh tokens have foreign keys with `ON DELETE CASCADE`, their associated records are deleted as well.

The endpoint only works when:

```env
PLATFORM=dev
```

**Warning:** This is a destructive development endpoint and should not be exposed in a production environment.

## Static Files and Metrics

The server exposes the `fileserver` directory under `/app/`:

```http
GET /app/
GET /app/assets/logo.png
```

Requests to `/app/` pass through a metrics middleware that increments an atomic hit counter. The current count can be viewed through `/admin/metrics`.

## Authentication

Chirpy uses two token types.

### Access Token

The access token is a JWT signed using `TOKEN_SECRET`.

- Token type/issuer: `chirpy-access`
- Lifetime: 1 hour
- Contains the authenticated user's UUID as the JWT subject
- Used with protected API endpoints

Example:

```http
Authorization: Bearer <access-token>
```

### Refresh Token

Refresh tokens are randomly generated, stored in PostgreSQL, and associated with a user.

- Lifetime: 60 days
- Can be revoked
- Used to obtain a new access token without logging in again

Example:

```http
Authorization: Bearer <refresh-token>
```

### Password Security

Passwords are never stored directly. They are hashed and verified using Argon2id.

## Database

Chirpy uses PostgreSQL for persistent storage.

The current database consists of:

- `users` — user accounts, password hashes, and Chirpy Red membership status.
- `chirps` — chirp content and its associated author.
- `refresh_tokens` — refresh tokens, expiration timestamps, and revocation timestamps.

Users are referenced by chirps and refresh tokens with foreign keys using `ON DELETE CASCADE`.

### Database Files

Schema migrations:

```text
internal/sql/schema/
```

SQL queries:

```text
internal/sql/queries/
```

Generated database-access code:

```text
internal/database/
```

## sqlc

The project uses [sqlc](https://sqlc.dev/) to generate type-safe Go database code from SQL queries.

The configuration is defined in `sqlc.yaml`:

```yaml
version: "2"
sql:
  - schema: "internal/sql/schema"
    queries: "internal/sql/queries"
    engine: "postgresql"
    gen:
      go:
        out: "internal/database"
```

After modifying SQL queries or the database schema, regenerate the Go database layer with:

```bash
sqlc generate
```

## Architecture

The application is organized around a few main layers:

```text
Client
  │
  ▼
main.go
  │
  ▼
HTTP routes and handlers
  │
  ├── User operations
  ├── Chirp operations
  ├── Authentication
  ├── Admin operations
  └── Polka webhook
  │
  ├───────────────┐
  ▼               ▼
internal/auth   internal/database
  │               │
  ├── JWT         ├── sqlc-generated queries
  └── Argon2id    └── database models
                  │
                  ▼
              PostgreSQL
```

`main.go` constructs the application configuration and registers the HTTP routes. The handlers coordinate HTTP requests, authentication, application operations, and database access. The `internal/auth` package contains JWT, bearer-token, API-key, refresh-token, and password-hashing functionality. The `internal/database` package contains sqlc-generated database code.

## Project Structure

```text
chirpy/
├── cmd/
│   ├── handlers.go
│   ├── helpers.go
│   ├── helpers_test.go
│   └── structs.go
├── fileserver/
│   ├── assets/
│   │   └── logo.png
│   └── index.html
├── internal/
│   ├── auth/
│   │   ├── auth.go
│   │   ├── auth_test.go
│   │   └── hash.go
│   ├── database/
│   │   ├── chirps.sql.go
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── refresh_tokens.sql.go
│   │   └── users.sql.go
│   └── sql/
│       ├── queries/
│       │   ├── chirps.sql
│       │   ├── refresh_tokens.sql
│       │   └── users.sql
│       └── schema/
│           ├── 001_users.sql
│           ├── 002_chirps.sql
│           ├── 003_user_password.sql
│           ├── 004_revoked_at.sql
│           └── 005_chirpy_red_membership.sql
├── main.go
├── request.rest
├── sqlc.yaml
├── go.mod
└── go.sum
```

## Testing

Run all tests with:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

Tests currently cover helper functionality and authentication functionality, including JWT and password-related behavior.

## Example Workflow

A typical API workflow is:

1. Start PostgreSQL.
2. Create the `chirpy` database.
3. Configure `.env` with `DB_URL`, `PLATFORM`, `TOKEN_SECRET`, and `POLKA_KEY`.
4. Run the Goose migrations.
5. Start the server with `go run .`.
6. Register a user with `POST /api/users`.
7. Log in with `POST /api/login`.
8. Use the returned access token for protected endpoints.
9. Create and retrieve chirps.
10. Filter chirps by author or sort them in descending order when needed.
11. Use the refresh token to obtain a new access token.
12. Revoke the refresh token when it should no longer be usable.
13. Use the authenticated Polka webhook to upgrade a user to Chirpy Red.

The repository also contains `request.rest` with example HTTP requests that can be used with REST clients such as the VS Code REST Client extension.

## Tech Stack

- Go 1.27
- `net/http`
- PostgreSQL
- SQL
- sqlc
- Goose migrations
- JWT (`github.com/golang-jwt/jwt/v5`)
- Argon2id (`github.com/alexedwards/argon2id`)
- UUID (`github.com/google/uuid`)
- `godotenv`
- `sync/atomic` for file-server metrics

## Learning Goals

This project focuses on practical backend development concepts in Go, including:

- Building HTTP servers with `net/http`
- HTTP routing and handlers
- REST-style API design
- Request and response handling
- Middleware
- Authentication and authorization
- JWT access tokens
- Refresh-token lifecycle management
- Password hashing with Argon2id
- API-key authenticated webhooks
- PostgreSQL integration
- SQL schema design
- Database migrations
- Type-safe SQL code generation with sqlc
- UUIDs and foreign-key relationships
- Cascading deletes
- Atomic counters and middleware state
- Testing Go code
- Environment-based configuration

## License

This project is currently a learning project and does not specify a separate open-source license.
