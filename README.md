# Go Web App

Basic clean structure for a Go web application with Gin.

## Run

```bash
go run ./cmd
```

The server starts on `:8080` by default.

## Structure

- `cmd` - application entry point
- `internal/api` - Gin router and handlers
- `internal/domain` - domain models
- `internal/service` - business logic
- `internal/repository` - data access layer
- `database` - PostgreSQL connection
