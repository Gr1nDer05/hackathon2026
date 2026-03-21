# Module Boundaries (Simplified)

## Dependency Rule

Allowed direction:

`app -> pages -> modules -> shared`

Rules:
- `pages` import business logic only from `modules/*`.
- `modules` may use `shared/*`, but never `pages/*` or `app/*`.
- API calls are concentrated inside `modules/*/api` or `shared/api` helpers.

## Module Responsibilities

### `modules/auth`
- login request
- token cookie lifecycle
- current user bootstrap (`/auth/me`)
- subscription status normalization for initial access decision

### `modules/subscription` (planned)
- unified subscription policy checks
- reusable guards/helpers for paid access rules

### `modules/profile` (planned)
- profile fetch/update for admin and psychologist contexts

## Pages Ownership

- `pages/auth/*`: sign-in entry.
- `pages/psychologist/*`: dashboard/tests/builder/results/profile.
- `pages/admin/*`: psychologists management and subscriptions.
- `pages/client/*`: token session and final result.
- `pages/common/*`: forbidden, not-found, subscription-required.

## Router Ownership

- Only router layer composes guards and redirects.
- Route constants come only from `shared/config/routes.js`.
