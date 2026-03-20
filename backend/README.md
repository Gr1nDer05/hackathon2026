# Go Web App

Basic clean structure for a Go web application with Gin.

## Run Locally

```bash
cp .env.example .env
go run ./cmd
```

The server starts on `:8080` by default.
On startup the app automatically applies all `migrations/*.up.sql` files to PostgreSQL.

## Publish To Docker Hub

```bash
docker build -t gr1nder05/hackathon2026-backend:latest .
docker login
docker push gr1nder05/hackathon2026-backend:latest
```

## Run With Docker

```bash
cp .env.example .env
docker compose up
```

The backend will be available on `http://localhost:8080`.

The `app` service is built locally, applies migrations on startup, and PostgreSQL is started automatically from the official image.

## Structure

- `cmd` - application entry point
- `internal/api` - Gin router and handlers
- `internal/domain` - domain models
- `internal/service` - business logic
- `internal/repository` - data access layer
- `database` - PostgreSQL connection
