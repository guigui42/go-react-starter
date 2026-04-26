# Architecture

## Overview

Go React Starter follows a layered architecture with clear separation of concerns.

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (React)                     │
│  Vite + React 19 + TanStack Router + TanStack Query     │
│  Mantine UI + i18n (EN/FR) + PWA                        │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP/JSON (port 5173 → 8080)
┌──────────────────────▼──────────────────────────────────┐
│                    Backend (Go/Chi)                      │
│  Middleware → Handlers → Services → Repositories         │
│  JWT auth + CSRF + Rate Limiting + Audit Logging         │
└──────────────────────┬──────────────────────────────────┘
                       │ GORM (PostgreSQL driver)
┌──────────────────────▼──────────────────────────────────┐
│                  PostgreSQL 17                           │
│  Row-level security via GORM scopes                      │
└─────────────────────────────────────────────────────────┘
```

## Backend Layers

### Request Flow

```
HTTP Request
  → chi.Router (route matching)
    → Middleware Stack (auth, CSRF, rate limit, logging)
      → Handler (HTTP concerns: parse request, write response)
        → Service (business logic, validation)
          → Repository (data access via GORM)
            → PostgreSQL
```

### Middleware Stack (order matters)

| Order | Middleware | Purpose |
|-------|-----------|---------|
| 1 | RequestID | Assigns unique ID to each request |
| 2 | RealIP | Extracts client IP from proxy headers |
| 3 | OTel Route Labeler | OpenTelemetry span naming (if enabled) |
| 4 | Logging | Request/response logging with Zerolog |
| 5 | Recovery | Panic recovery → JSON error response |
| 6 | CORS | Cross-origin headers (dev mode) |

Route-level middleware (applied per-group):
- **Auth middleware** — JWT validation, user context injection
- **CSRF middleware** — Double-submit cookie verification
- **Admin middleware** — `is_admin` flag check
- **Rate limiting** — IP-based and per-user limits
- **Body size** — Max request body enforcement

### Handler Layer

Handlers deal exclusively with HTTP:
- Parse request (query params, body, path params)
- Call service methods
- Write JSON responses via `response.Success()` / `response.Error()`
- Never contain business logic

### Service Layer

Services contain business logic:
- Input validation beyond HTTP parsing
- Orchestration of multiple repositories
- Audit event logging
- Error wrapping with domain context

### Repository Layer

Repositories handle data access:
- GORM queries with user scoping (`scopes.ForUser(userID)`)
- Transaction management
- No HTTP or business logic awareness

## Frontend Architecture

### Feature-Based Structure

Each feature is a self-contained module:

```
features/
├── auth/              # Authentication flows
│   ├── components/    # LoginForm, RegisterForm, OAuthButtons, ...
│   ├── hooks/         # useAuth, usePasskeys, useOAuth, ...
│   └── types/         # TypeScript interfaces
├── admin/             # Admin dashboard
│   ├── components/    # StatsCards, UserList, AuditLogViewer, ...
│   ├── hooks/         # useAdminStats, useAuditLogs, ...
│   └── types/         # Admin-specific types
├── notes/             # Sample CRUD feature
│   ├── components/    # NoteCard, NoteForm, NoteList
│   ├── hooks/         # useNotes, useCreateNote, ...
│   └── types/         # Note interfaces
└── user/              # User settings
    ├── components/    # ProfileSettings, PrivacySettings, ...
    └── hooks/         # usePreferences, useDeleteAccount, ...
```

### Routing

TanStack Router with file-based routing:

| Route | File | Auth | Description |
|-------|------|------|-------------|
| `/` | `routes/index.tsx` | No | Landing page |
| `/login` | `routes/login.tsx` | No | Login page |
| `/register` | `routes/register.tsx` | No | Registration |
| `/settings` | `routes/settings.tsx` | Yes | User settings |
| `/notes` | `routes/notes/index.tsx` | Yes | Notes list |
| `/notes/:noteId` | `routes/notes/$noteId.tsx` | Yes | Note detail |
| `/admin` | `routes/admin/index.tsx` | Admin | Admin dashboard |

### State Management

- **Server state**: TanStack Query (caching, mutations, optimistic updates)
- **Auth state**: React context + httpOnly cookie (JWT)
- **UI state**: Component-local `useState` / Mantine hooks
- **No global state library** — TanStack Query covers most needs

## Production Architecture

```
┌─────────────────────────────────────────┐
│            Docker Container              │
│                                          │
│  supervisord                             │
│  ├── nginx (port 80)                     │
│  │   ├── /           → static SPA files  │
│  │   ├── /api/*      → proxy :8080       │
│  │   ├── /auth/*     → proxy :8080       │
│  │   ├── /health     → proxy :8080       │
│  │   └── /metrics    → proxy :8080       │
│  └── backend (port 8080)                 │
│      └── Go binary                       │
└─────────────────────────────────────────┘
```

nginx handles:
- Static file serving with gzip compression
- Reverse proxy to the Go API
- Security headers (CSP, X-Frame-Options, etc.)
- SPA fallback (all non-API routes → `index.html`)
