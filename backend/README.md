# Hackathon2026 Backend

Backend on Go + Gin + PostgreSQL.

On startup the application automatically:

- connects to PostgreSQL
- applies all migrations from `migrations/*.up.sql`
- seeds admin accounts from `ADMIN_ACCOUNTS`
- seeds demo psychologist, report template, and demo test when `DEMO_DATA_ENABLED=true`
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

To inspect PostgreSQL in the browser during local development:

```bash
docker compose up -d db adminer
```

Then open `http://localhost:8081` and use:

- `System`: `PostgreSQL`
- `Server`: `db`
- `Username`: `DB_USER` from `.env`
- `Password`: `DB_PASSWORD` from `.env`
- `Database`: `DB_NAME` from `.env`

## Push Backend Image To Docker Hub

Replace `gr1nder05` with your Docker Hub username if needed.

1. Login:

```bash
docker login
```

2. Build image:

```bash
docker build -t gr1nder05/hackathon2026-backend:latest .
```

3. Push image:

```bash
docker push gr1nder05/hackathon2026-backend:latest
```

Optional version tag:

```bash
docker build -t gr1nder05/hackathon2026-backend:v1 .
docker push gr1nder05/hackathon2026-backend:v1
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
APP_IMAGE=gr1nder05/hackathon2026-backend:latest
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
- `DEMO_DATA_ENABLED` - enables demo psychologist, demo report template, and demo test seeding
- `DEMO_PSYCHOLOGIST_EMAIL`, `DEMO_PSYCHOLOGIST_PASSWORD`, `DEMO_PSYCHOLOGIST_FULL_NAME` - demo workspace credentials
- `ADMINER_PORT` - browser port for Adminer in local development

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

## Architecture

The backend follows a simple layered structure:

- `internal/api` handles HTTP transport, cookie auth, CSRF, validation, and HTTP response shaping.
- `internal/service` contains product logic: workspace flows, test constructor behavior, public completion, career scoring, report generation, and demo seeding.
- `internal/repository` contains SQL queries and keeps PostgreSQL details isolated from handlers and services.
- `database` manages connections and migration execution.

Key technical decisions:

- Go + Gin without ORM. SQL stays explicit, which keeps aggregate queries, migrations, and reporting logic predictable.
- Cookie sessions + CSRF protection. This fits the browser-first frontend flow and keeps auth simple for both admin and psychologist кабинеты.
- Reports are generated on demand. The server first renders HTML, then builds DOCX from that HTML, so report files are never stored on disk and both formats stay consistent.
- Report templates are stored in PostgreSQL as JSON configs. Tests reference them through `report_template_id`, and templates can customize client and psychologist reports separately.
- Public test submissions store a snapshot of `career_result`, so historical reports do not change retroactively after later edits to a test.
- Demo data is seeded automatically in non-production mode by default, which gives the team a ready-made workspace for frontend and demo work.

## Demo Data

When `DEMO_DATA_ENABLED=true`, startup seeds:

- a demo psychologist account
- a demo report template
- a published demo test `ПрофДНК: IT-профориентация`
- five scale-based demo questions connected to the career guidance engine

Default demo psychologist credentials:

- email: `demo.psychologist@profdnk.local`
- password: `demo12345`

This makes the “case requirement” about a pre-created test and example report templates available immediately after startup.
