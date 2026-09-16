# Chirpy

Chirpy is a backend HTTP API for a small social-media-style application built with Go. Users can create accounts, log in, create and view short posts (chirps), and manage authentication with JWT access tokens and refresh tokens.

The project was built as a practical exercise in building a Go web server, working with PostgreSQL, SQL migrations, authentication, password hashing, middleware, and REST-style API endpoints.

## Features

- User registration and account updates
- Password hashing with Argon2id
- User login with JWT access tokens
- Refresh-token based authentication
- Refresh-token revocation
- Create, list, retrieve, and delete chirps
- Chirp profanity filtering
- Protected endpoints using bearer-token authentication
- PostgreSQL persistence
- SQL queries generated with sqlc
- Database schema migrations using Goose-compatible SQL files
- Static file serving for the application
- Request/file-server metrics
- Development-only database reset endpoint

## Prerequisites

You will need:

- Go 1.27 or later
- PostgreSQL
- Goose for applying the database migrations

The project uses Go 1.27.0 and depends on PostgreSQL, JWT, Argon2id, UUID, and environment-variable packages.

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/Paa-Kwasi-04/chirpy.git
cd chirpy
```

### 2. Install dependencies

Download the Go dependencies:

```bash
go mod download
```

### 3. Create the PostgreSQL database

Create a PostgreSQL database for Chirpy. For example:

```sql
CREATE DATABASE chirpy;
```

Your database connection string will generally look like:

```text
postgres://username:password@localhost:5432/chirpy?sslmode=disable
```

### 4. Configure environment variables

Chirpy loads configuration from environment variables using a `.env` file.

Create a `.env` file in the project root:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
TOKEN_SECRET=replace-with-a-random-secret
```

- `DB_URL` is the PostgreSQL connection string.
- `PLATFORM` controls the environment. The database reset endpoint is only available when this is set to `dev`.
- `TOKEN_SECRET` is used to sign and validate JWT access tokens.

Do not commit `.env` or real secrets to Git.

### 5. Run the database migrations

The schema files are located in `internal/sql/schema` and use Goose migration annotations.

Install Goose if you do not already have it:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Then apply the migrations:

```bash
goose -dir internal/sql/schema postgres "$DB_URL" up
```

On Windows PowerShell, you can use the connection string directly if the environment variable is not expanded as expected:

```powershell
goose -dir internal/sql/schema postgres "postgres://username:password@localhost:5432/chirpy?sslmode=disable" up
```

## Running Chirpy

Start the server with:

```bash
go run .
```

The API listens on:

```text
http://localhost:8080
```

You can also build the application:

```bash
go build -o chirpy .
```

Then run the resulting executable.

## API Endpoints

### Health check

```http
GET /api/healthz
```

Returns a successful response when the server is running.

### Create a user

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

### Update a user

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

A successful login returns an access token and refresh token.

### Refresh an access token

```http
POST /api/refresh
Authorization: Bearer <refresh-token>
```

Returns a new JWT access token when the refresh token is valid and has not expired or been revoked.

### Revoke a refresh token

```http
POST /api/revoke
Authorization: Bearer <refresh-token>
```

Revokes the supplied refresh token.

### Create a chirp

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

Chirps are limited to 140 characters. The application also replaces configured profane words before storing the chirp.

### Get all chirps

```http
GET /api/chirps
```

Returns the chirps stored in the database.

### Get a single chirp

```http
GET /api/chirps/{chirpID}
```

Replace `{chirpID}` with the chirp's UUID.

### Delete a chirp

```http
DELETE /api/chirps/{chirpID}
Authorization: Bearer <access-token>
```

Only the user who owns the chirp can delete it.

## Admin Endpoints

### Metrics

```http
GET /admin/metrics
```

Returns a small HTML page showing the number of visits to the `/app/` file server.

### Reset

```http
POST /admin/reset
```

Deletes all users and, through the database relationships, their associated chirps and refresh tokens.

This endpoint only works when:

```env
PLATFORM=dev
```

**Warning:** This is a destructive development endpoint. Do not expose it in a production environment.

## Static Files

The `/app/` route serves files from the `fileserver` directory.

```text
GET /app/
```

Requests to this route also increment the application's file-server hit counter.

## Authentication

Chirpy uses two types of tokens:

- **Access token:** a JWT signed with the configured `TOKEN_SECRET`. Access tokens expire after 1 hour.
- **Refresh token:** a random token stored in PostgreSQL and associated with a user. Refresh tokens expire after 60 days and can be revoked.

Protected endpoints expect the access token in the `Authorization` header:

```text
Authorization: Bearer <access-token>
```

Passwords are hashed using Argon2id before being stored in the database.

## Database

Chirpy uses PostgreSQL for persistent storage.

The main tables are:

- `users` — stores user accounts and password hashes.
- `chirps` — stores chirps and associates each chirp with its author.
- `refresh_tokens` — stores refresh tokens, expiration times, and revocation timestamps.

The SQL schema is stored in:

```text
internal/sql/schema/
```

SQL queries are stored in:

```text
internal/sql/queries/
```

Generated database-access code is stored in:

```text
internal/database/
```

## sqlc

The project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

The configuration is defined in `sqlc.yaml`.

After changing the SQL schema or queries, regenerate the database code with:

```bash
sqlc generate
```

## Testing

Run the Go test suite with:

```bash
go test ./...
```

For more detailed test output:

```bash
go test -v ./...
```

## Example Workflow

A basic workflow for using the API is:

1. Start PostgreSQL.
2. Create the `chirpy` database.
3. Configure `.env`.
4. Run the Goose migrations.
5. Start the server with `go run .`.
6. Create a user with `POST /api/users`.
7. Log in with `POST /api/login`.
8. Use the returned access token to create chirps.
9. Retrieve chirps with `GET /api/chirps`.
10. Use the refresh token to obtain a new access token when necessary.
11. Revoke the refresh token when it should no longer be usable.

## Tech Stack

- Go
- `net/http`
- PostgreSQL
- SQL
- sqlc
- Goose migrations
- JWT (`golang-jwt/jwt`)
- Argon2id password hashing
- UUID
- `godotenv`

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
│   └── index.html
├── internal/
│   ├── auth/
│   │   ├── auth.go
│   │   ├── auth_test.go
│   │   └── hash.go
│   ├── database/
│   └── sql/
│       ├── queries/
│       └── schema/
├── main.go
├── request.rest
├── sqlc.yaml
├── go.mod
└── go.sum
```

## Learning Goals

This project focuses on practical backend development concepts in Go, including:

- Building HTTP servers with `net/http`
- Designing REST-style API endpoints
- HTTP request/response handling
- Middleware
- Authentication and authorization
- JWTs and refresh tokens
- Secure password hashing
- PostgreSQL database integration
- SQL schema design
- SQL migrations
- Type-safe SQL code generation with sqlc
- Working with UUIDs
- Testing Go code
- Environment-based configuration

## License

This project is primarily a learning project and does not currently specify a separate open-source license.
