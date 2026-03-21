# Frontend Architecture (Simplified)

## Goal

Keep the project easy to grow and easy to read while covering key case flows: auth, psychologist workspace, admin panel, and client test session.

## Core Constraints

- Psychologists are created by admin, no self-registration.
- Psychologist work routes require active paid subscription.
- Client goes through tokenized public flow.

## Folder Structure

```text
src/
  app/                  # App root, providers, router bootstrap
  pages/                # role-oriented screens
    auth/
    psychologist/
    admin/
    client/
    common/
  modules/              # business logic modules
    auth/               # api, cookie/token, hooks, login ui
    profile/            # planned
    subscription/       # planned
  shared/               # config, shared api client, ui primitives, utils
```

## Routing Ownership

- Path constants: `src/shared/config/routes.js`
- Route composition + guards: `src/app/router` (planned)
- Pages do not contain routing rules for other pages.

## Access Model

- `RequireAuth`: authenticated-only routes.
- `RequireRole`: role gate (`admin` / `psychologist`).
- `RequireActiveSubscription`: psychologist work routes only.

Guard chain for psychologist workspace:
1. `RequireAuth`
2. `RequireRole("psychologist")`
3. `RequireActiveSubscription`

## Data Contract (minimum)

`GET /auth/me` returns:
- `id`
- `role` (`admin` | `psychologist`)
- `firstName`, `lastName`, `username`
- `subscriptionStatus` (`active` | `expired` | `blocked` | `trial`)

Compatibility aliases allowed in mapper:
- `subscription.status`
- `subscription_state`
- `tariffStatus`

## Current Status

Implemented:
- `modules/auth` as single auth source.
- `pages/auth/AuthPage.jsx` consuming only `modules/auth`.

Planned next:
- router file with route matrix from `docs/architecture/ROUTES_AND_GUARDS.md`
- page shells for admin/psychologist/client/common
