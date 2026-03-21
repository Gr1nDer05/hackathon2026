# Routes And Guards Matrix

## Route ownership

| Route | Audience | Guards | Page group |
|---|---|---|---|
| `/` | public | none | `pages/auth` |
| `/session/:token` | client | none | `pages/client` |
| `/result/:token` | client | none | `pages/client` |
| `/dashboard` | psychologist | `RequireAuth` + `RequireRole(psychologist)` + `RequireActiveSubscription` | `pages/psychologist` |
| `/tests` | psychologist | same as above | `pages/psychologist` |
| `/builder/:id` | psychologist | same as above | `pages/psychologist` |
| `/tests/:id/results` | psychologist | same as above | `pages/psychologist` |
| `/profile` | psychologist | `RequireAuth` + `RequireRole(psychologist)` | `pages/psychologist` |
| `/admin` | admin | `RequireAuth` + `RequireRole(admin)` | `pages/admin` |
| `/admin/psychologists` | admin | `RequireAuth` + `RequireRole(admin)` | `pages/admin` |
| `/admin/psychologists/:id` | admin | `RequireAuth` + `RequireRole(admin)` | `pages/admin` |
| `/admin/subscriptions` | admin | `RequireAuth` + `RequireRole(admin)` | `pages/admin` |
| `/subscription-required` | psychologist | `RequireAuth` + `RequireRole(psychologist)` | `pages/common` |
| `/forbidden` | protected | `RequireAuth` | `pages/common` |
| `*` | public | none | `pages/common` |

## Redirect rules

- Unauthenticated access to protected route -> `/`.
- Authenticated psychologist without active subscription trying to open work route -> `/subscription-required`.
- Authenticated user with wrong role -> `/forbidden`.
- Authenticated user opening `/` -> role-based default page:
  - `admin` -> `/admin`
  - `psychologist` with active subscription -> `/dashboard`
  - `psychologist` without active subscription -> `/subscription-required`
