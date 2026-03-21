# Hackathon2026 Backend

Backend on Go + Gin + PostgreSQL.

On startup the application automatically:

- connects to PostgreSQL
- applies all migrations from `migrations/*.up.sql`
- seeds admin accounts from `ADMIN_ACCOUNTS`
- starts HTTP server on port `8080`

## Local Run Without Docker

1. Copy env file:

```bash
cp .env.example .env
```

2. Make sure PostgreSQL is running locally.

3. Start backend:

```bash
go run ./cmd
```

Backend will be available at `http://localhost:8080`.

## Local Run With Docker Compose

This option builds the backend image locally and starts both backend and database.

1. Copy env file:

```bash
cp .env.example .env
```

2. Start services:

```bash
docker compose up --build -d
```

3. Check logs:

```bash
docker compose logs -f app
```

4. Stop services:

```bash
docker compose down
```

If you also want to delete PostgreSQL data volume:

```bash
docker compose down -v
```

## Push Backend Image To Docker Hub

Replace `gr1nder005` with your Docker Hub username if needed.

1. Login:

```bash
docker login
```

2. Build image:

```bash
docker build -t gr1nder005/hackathon2026-backend:latest .
```

3. Push image:

```bash
docker push gr1nder005/hackathon2026-backend:latest
```

Optional version tag:

```bash
docker build -t gr1nder005/hackathon2026-backend:v1 .
docker push gr1nder005/hackathon2026-backend:v1
```

## Run Pulled Image From Docker Hub

File `docker-compose.deploy.yml` is for teammates. It does not build locally and instead pulls the backend image from Docker Hub.

### What To Send To Frontend Teammate

Send them:

- `docker-compose.deploy.yml`
- `.env.example`
- image name in Docker Hub, for example `gr1nder05/hackathon2026-backend:latest`

### What Frontend Teammate Should Do

1. Create a folder and place `docker-compose.deploy.yml` there.

2. Create `.env` in the same folder as the compose file. Minimal example:

```env
APP_IMAGE=gr1nder005/hackathon2026-backend:latest
APP_ENV=development
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=postgres
COOKIE_SECURE=false
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
PUBLIC_BASE_URL=http://localhost:8080
ADMIN_ACCOUNTS=admin:change_me_please:Platform Administrator
```

3. Start backend and database:

```bash
docker compose -f docker-compose.deploy.yml up -d
```

4. If image was updated and they need the newest version:

```bash
docker compose -f docker-compose.deploy.yml pull
docker compose -f docker-compose.deploy.yml up -d
```

5. Check logs:

```bash
docker compose -f docker-compose.deploy.yml logs -f app
```

6. Stop everything:

```bash
docker compose -f docker-compose.deploy.yml down
```

7. Stop and remove database data too:

```bash
docker compose -f docker-compose.deploy.yml down -v
```

## Important Env Variables

- `APP_IMAGE` - Docker image name for deploy compose
- `DB_NAME`, `DB_USER`, `DB_PASSWORD` - PostgreSQL credentials
- `ALLOWED_ORIGINS` - frontend URLs allowed by CORS
- `PUBLIC_BASE_URL` - public backend URL used when generating public links
- `ADMIN_ACCOUNTS` - admin seed in format `login:password:Full Name`
- `COOKIE_SECURE` - set `true` only when backend works over HTTPS

## Swagger / OpenAPI

After startup:

- `http://localhost:8080/swagger/`
- `http://localhost:8080/swagger/openapi.yaml`

## Project Structure

- `cmd` - application entry point
- `internal/api` - Gin router and handlers
- `internal/domain` - domain models
- `internal/service` - business logic
- `internal/repository` - data access layer
- `database` - PostgreSQL connection and migrations
