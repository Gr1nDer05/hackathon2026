# Backend Launch Guide For Frontend

This guide explains how to run the backend and PostgreSQL locally from a ready Docker image.

## What You Need

- Docker
- Docker Compose
- files `docker-compose.deploy.yml` and `.env`

## Files

Place these files in one folder:

- `docker-compose.deploy.yml`
- `.env`

If you only received `.env.example`, rename it to `.env`.

## .env Example

Use the following values:

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
DEMO_DATA_ENABLED=true
OPENAI_API_KEY=
OPENAI_REPORT_TEMPLATE_MODEL=gpt-5-mini
```

## Start

Run:

```bash
docker compose -f docker-compose.deploy.yml up -d
```

This will automatically start:

- backend
- PostgreSQL
- Adminer on `http://localhost:8081`

## Check That Everything Works

Backend logs:

```bash
docker compose -f docker-compose.deploy.yml logs -f app
```

Backend URL:

- `http://localhost:8080`

Swagger URL:

- `http://localhost:8080/swagger/`

Adminer URL:

- `http://localhost:8081`

Demo psychologist credentials:

- email: `demo.psychologist@profdnk.local`
- password: `demo12345`

Demo psychologist with expired subscription:

- email: `expired.psychologist@profdnk.local`
- password: `expired12345`

## Browser / Auth Notes

- admin login: `POST /auth/admin/login`
- psychologist login: `POST /auth/psychologists/login`
- the backend uses session cookies, not bearer tokens
- login sets a session cookie and `csrf_token`
- for protected non-`GET` requests, send header `X-CSRF-Token` with the same value as the `csrf_token` cookie
- direct browser requests are supported when the frontend origin is listed in `ALLOWED_ORIGINS`
- preflight `OPTIONS` requests return CORS headers for `Content-Type` and `X-CSRF-Token`
- API errors use `{ "message": "...", "field_errors": { ... } }`
- report templates are managed through `/psychologists/report-templates`
- `POST /psychologists/report-templates/generate` creates an AI draft from one prompt; it is available only for psychologists with `subscription_plan=pro`
- `POST /psychologists/me/subscription/purchase` is a placeholder "buy subscription" action from the psychologist cabinet; the psychologist chooses `basic` or `pro`, and the backend creates a pending 30-day purchase request
- `GET /admins/me/subscription-purchase-requests` returns pending purchase requests so the admin UI can show them as manual subscription notifications
- `GET /psychologists/tests` returns extra dashboard metrics such as started/completed counts and last activity timestamps
- `GET /public/tests/{slug}` includes a stable `psychologist.user/profile/card` block for the public author page
- tests support `show_client_report_immediately`; when it is `true`, completed public submissions return `client_report_available` and `client_report_url`
- clients can open their own report through `GET /public/tests/{slug}/report?access_token=...&format=html|docx`
- completed test results return both legacy `career_result` and universal `metrics` / `top_metrics`
- psychologist user payloads now include `subscription_plan` (`basic` or `pro`)
- admin access updates accept `subscription_plan` in `PUT /admins/psychologists/{id}/access`
- placeholder purchase requests do not activate access automatically; they only create an admin-side notification record

## Test Link Access Mode

When creating or updating a test through the API:

- send `has_participant_limit: true` and a positive `max_participants` to limit the number of people who can open the public link
- send `has_participant_limit: false` to make the public link available to anyone
- legacy behavior is still supported: `max_participants: 0` also means the link is available to anyone

## Stop

Stop containers:

```bash
docker compose -f docker-compose.deploy.yml down
```

Stop containers and remove database volume:

```bash
docker compose -f docker-compose.deploy.yml down -v
```

## Update To New Backend Version

If a new image was pushed to Docker Hub, run:

```bash
docker compose -f docker-compose.deploy.yml pull
docker compose -f docker-compose.deploy.yml up -d
```

## Notes

- no local backend build is needed
- database starts automatically
- if the frontend runs on another port, update `ALLOWED_ORIGINS` in `.env`
- Swagger is the source of truth for exact request and response schemas
- if `OPENAI_API_KEY` is not configured, AI template draft generation returns `503 Service Unavailable`
