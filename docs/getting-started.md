# Getting Started

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Go | 1.24+ | `go version` |
| Node.js | 25+ | `node --version` |
| Docker | 20+ | `docker --version` |
| Make | any | `make --version` |

## Quick Start (3 commands)

```bash
make install    # Install Go + npm dependencies
make db-init    # Start PostgreSQL + run migrations + seed admin
make dev        # Start backend (8080) + frontend (5173)
```

Open http://localhost:5173 in your browser.

## Step-by-Step Setup

### 1. Clone and Install Dependencies

```bash
git clone <your-repo-url>
cd go-react-starter
make install
```

This runs `go mod download` in the backend and `npm install` in the frontend.

### 2. Start the Database

```bash
make db-up      # Start PostgreSQL container (port 5434)
make db-wait    # Wait until PostgreSQL is accepting connections
```

Or use `make db-init` which does both + runs migrations.

The default port is **5434** to avoid conflicts. Override with:

```bash
DB_PORT=5432 make db-up
```

### 3. Configure Environment

```bash
cp backend/.env.example backend/.env
```

The `.env.example` has sensible development defaults. For production, you **must** set:

| Variable | Required | Purpose |
|----------|----------|---------|
| `JWT_SECRET` | ✅ | Token signing (min 32 chars) |
| `ENCRYPTION_KEY` | ✅ | Data encryption key |
| `PGPASSWORD` | ✅ | Database password |

### 4. Start Development Servers

```bash
make dev
```

This starts both servers in parallel:
- **Backend**: http://localhost:8080 (Go/Chi API)
- **Frontend**: http://localhost:5173 (Vite dev server with HMR)

### 5. Create an Admin Account

Register at http://localhost:5173/register with the email configured in `ADMIN_EMAILS` (default: `admin@example.com`). The account is automatically promoted to admin.

## Makefile Reference

### Development

| Command | Description |
|---------|-------------|
| `make dev` | Start backend + frontend in parallel |
| `make dev-backend` | Start backend only |
| `make dev-frontend` | Start frontend only |
| `make install` | Install all dependencies |
| `make update` | Update all dependencies |

### Database

| Command | Description |
|---------|-------------|
| `make db-up` | Start PostgreSQL container |
| `make db-down` | Stop PostgreSQL container |
| `make db-restart` | Restart PostgreSQL |
| `make db-init` | Start DB + wait + ready |
| `make db-reset` | Destroy and recreate database |
| `make db-shell` | Open psql shell |
| `make db-logs` | Tail PostgreSQL logs |

### Testing & Quality

| Command | Description |
|---------|-------------|
| `make test` | Run all tests (backend + frontend) |
| `make test-backend` | Run Go tests |
| `make test-frontend` | Run Vitest tests |
| `make lint` | Lint all code |
| `make fmt` | Format all code |

### Production

| Command | Description |
|---------|-------------|
| `make release` | Build Docker image |
| `make clean` | Remove build artifacts |

## Renaming the Project

When starting a new project, update these references:

1. **Go module**: `backend/go.mod` — change `github.com/example/go-react-starter`
2. **Package name**: `frontend/package.json` — change `name` field
3. **Docker image**: `docker/Dockerfile` and `docker-compose.yml`
4. **CI/CD**: `.github/workflows/*.yml` — image names, service names
5. **App title**: `frontend/index.html`, `frontend/src/i18n/locales/*.json`
6. **WebAuthn**: `backend/.env` — update `WEBAUTHN_RP_NAME`

Use a global find-and-replace:

```bash
# Replace module path
find backend -name '*.go' -exec sed -i 's|github.com/example/go-react-starter|github.com/yourorg/yourproject|g' {} +
sed -i 's|github.com/example/go-react-starter|github.com/yourorg/yourproject|g' backend/go.mod
```
