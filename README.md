# Go React Starter

A production-ready full-stack starter template with Go/Chi backend, React 19/TypeScript frontend, and PostgreSQL database. Extracted from a real production application — includes battle-tested authentication, admin panel, security features, and CI/CD pipelines.

## ✨ Features

### Authentication (All Methods)
- 📧 Email/password with bcrypt hashing
- 🔑 Passkey/WebAuthn (passwordless login)
- 🔐 OAuth (Google, GitHub, Facebook)
- 🛡️ Backup codes for account recovery
- ✉️ Email verification
- 🔒 Account lockout (brute-force protection)
- 🍪 CSRF protection (double-submit cookie)

### Admin Panel
- �� User management
- 📊 System statistics
- 🔍 Audit log viewer
- 🛡️ CSP violation monitoring
- 📧 Email configuration tester
- 📋 Migration history

### Security & Observability
- 🔐 JWT with httpOnly cookies
- 🚦 Rate limiting (IP + per-user)
- 📝 Audit logging
- 🛡️ CSP violation reporting & persistence
- 🔑 Token blocklist (logout invalidation)
- 📊 OpenTelemetry tracing & metrics
- 📈 Prometheus metrics endpoint
- 🔒 Row-level security (GORM scopes)

### Developer Experience
- 🧪 Test infrastructure (testcontainers, per-test DB schemas)
- 🌍 i18n (English + French)
- 🎨 Mantine UI components (WCAG 2.1 AA)
- ⚡ Vite + HMR
- 🔄 TanStack Router (file-based, code-split)
- 🔍 TanStack Query (caching, mutations)
- 📱 PWA support
- 🐳 Docker (multi-stage build)
- 🔄 GitHub Actions CI/CD

### Sample Feature: Notes
A complete CRUD example demonstrating the full-stack pattern:
- Backend: Handler → Repository → Model (with user scoping)
- Frontend: Route → Component → Hook → API client

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 25+ (see `.nvmrc`)
- PostgreSQL 17+
- Docker (optional, for containerized development)

### 1. Start the Database

```bash
cd docker && docker compose up -d
```

### 2. Configure Environment

```bash
cp backend/.env.example backend/.env
# Edit backend/.env with your settings (at minimum: JWT_SECRET, ENCRYPTION_KEY, PGPASSWORD)
```

### 3. Start Development Servers

```bash
# Terminal 1: Backend
cd backend && make run

# Terminal 2: Frontend
cd frontend && npm install && npm run dev
```

The app is available at http://localhost:5173 (frontend) and http://localhost:8080 (API).

## 📁 Project Structure

```
go-react-starter/
├── backend/                    # Go API server
│   ├── cmd/server/            # Entry point
│   ├── internal/
│   │   ├── config/            # App configuration
│   │   ├── handlers/          # HTTP handlers
│   │   ├── middleware/        # Auth, CSRF, CORS, rate limiting, etc.
│   │   ├── migrations/        # Versioned DB migrations
│   │   ├── models/            # GORM models
│   │   ├── repository/        # Data access layer
│   │   ├── services/          # Business logic
│   │   └── testutil/          # Test helpers
│   ├── pkg/                   # Shared utilities
│   └── templates/             # Email templates
├── frontend/                   # React application
│   ├── src/
│   │   ├── components/        # Shared UI components
│   │   ├── features/          # Feature modules
│   │   │   ├── auth/          # Authentication
│   │   │   ├── admin/         # Admin panel
│   │   │   ├── notes/         # Sample CRUD feature
│   │   │   └── user/          # User settings
│   │   ├── i18n/              # Internationalization
│   │   ├── lib/               # API client, query config
│   │   └── routes/            # File-based routing
│   └── tests/                 # Test setup
├── docker/                     # Docker & nginx config
└── .github/workflows/          # CI/CD pipelines
```

## 📖 Documentation

Comprehensive guides are in the [`docs/`](docs/) folder:

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Setup, prerequisites, Makefile reference |
| [Architecture](docs/architecture.md) | System design, layers, request flow |
| [Authentication](docs/authentication.md) | All auth methods, JWT, sessions |
| [Adding Features](docs/adding-features.md) | Step-by-step new feature guide |
| [Security](docs/security.md) | Middleware, CSRF, RLS, CSP |
| [Database & Migrations](docs/database.md) | PostgreSQL, GORM, migration system |
| [Deployment](docs/deployment.md) | Docker, nginx, CI/CD |
| [Configuration](docs/configuration.md) | All environment variables |
| [Testing](docs/testing.md) | Test infrastructure, per-test DB isolation |
| [Frontend](docs/frontend.md) | React, routing, i18n, components |

## 📄 License

MIT
