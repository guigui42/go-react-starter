# Deployment

## Docker Production Build

### Multi-Stage Dockerfile

The `docker/Dockerfile` uses a 3-stage build:

1. **Frontend build** — Node.js builds the React SPA
2. **Backend build** — Go compiles the API binary
3. **Runtime** — Alpine Linux runs nginx + Go binary via supervisord

```bash
# Build the image
docker build -f docker/Dockerfile -t go-react-starter .

# Run it
docker run -p 80:80 \
  -e JWT_SECRET=your-secret \
  -e ENCRYPTION_KEY=your-key \
  -e PGHOST=db-host \
  -e PGPASSWORD=db-pass \
  go-react-starter
```

### Runtime Architecture

```
supervisord (PID 1)
├── nginx (port 80)
│   ├── Static files (SPA)
│   ├── /api/* → proxy to :8080
│   ├── /auth/* → proxy to :8080
│   └── Security headers (CSP, HSTS, etc.)
└── backend (port 8080)
    └── Go API server
```

### Image Size

The final image is ~50MB thanks to:
- Alpine base
- Statically compiled Go binary
- Only production frontend assets (no node_modules)

## nginx Configuration

`docker/nginx.conf` handles:

- **SPA routing**: All non-API paths serve `index.html`
- **API proxy**: `/api/*`, `/auth/*`, `/health`, `/metrics` → backend
- **Compression**: gzip for HTML, CSS, JS, JSON
- **Caching**: Long-lived cache for static assets with content hashes
- **Security headers**: CSP, X-Frame-Options, X-Content-Type-Options
- **Rate limiting**: Connection limits per IP

## CI/CD Pipelines

### GitHub Actions Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `backend-tests.yml` | Push/PR to main (backend/**) | Go tests, vet, lint |
| `frontend-tests.yml` | Push/PR to main (frontend/**) | Vitest, type-check, lint |
| `codeql.yml` | Push/PR to main, weekly schedule | Security analysis |
| `docker-build-push.yml` | Push to main, version tags | Build and push Docker image |
| `release.yml` | Version tags (`v*`) | Create GitHub release |

### Backend Tests (`backend-tests.yml`)

- Runs on Ubuntu with PostgreSQL service container
- Executes `go test ./...` with race detector
- Runs `go vet` and linting

### Frontend Tests (`frontend-tests.yml`)

- Runs on Ubuntu with Node.js 25
- Executes Vitest unit tests
- Runs TypeScript type checking
- Runs ESLint

### Docker Build (`docker-build-push.yml`)

- Builds multi-platform image (amd64)
- Pushes to GitHub Container Registry (ghcr.io)
- Tags: `latest` for main, semantic version for tags

### Release (`release.yml`)

- Triggered by version tags (`v1.0.0`, `v2.1.3`, etc.)
- Creates GitHub release with changelog
- Attaches Docker image reference

## Environment Variables for Production

### Required

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | Token signing key (min 32 chars, keep secret) |
| `ENCRYPTION_KEY` | Data encryption key |
| `PGHOST` | PostgreSQL hostname |
| `PGPORT` | PostgreSQL port |
| `PGUSER` | PostgreSQL username |
| `PGPASSWORD` | PostgreSQL password |
| `PGDATABASE` | Database name |

### Recommended

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `development` | Set to `production` |
| `TRUST_PROXY` | `false` | Set to `true` behind a reverse proxy |
| `PGSSLMODE` | `disable` | Set to `require` for production |
| `ADMIN_EMAILS` | — | Comma-separated admin email addresses |

### Optional (OAuth)

| Variable | Description |
|----------|-------------|
| `OAUTH_GITHUB_CLIENT_ID` | GitHub OAuth app client ID |
| `OAUTH_GITHUB_CLIENT_SECRET` | GitHub OAuth app secret |
| `OAUTH_GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | Google OAuth secret |
| `OAUTH_CALLBACK_URL` | OAuth callback base URL |

### Optional (Email)

| Variable | Default | Description |
|----------|---------|-------------|
| `SMTP_HOST` | — | SMTP server hostname |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USERNAME` | — | SMTP username |
| `SMTP_PASSWORD` | — | SMTP password |
| `EMAIL_FROM` | — | Sender email address |
| `EMAIL_VERIFY_REQUIRED` | `false` | Require email verification |

## Health Checks

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `GET /health` | No | Basic liveness check |
| `GET /api/v1/monitoring/health` | Yes | Authenticated health check |
| `GET /metrics` | No* | Prometheus metrics (if enabled) |

*Configure nginx to restrict `/metrics` access in production.

## Scaling Considerations

- **Stateless backend**: JWT in cookies, no server-side sessions → horizontal scaling works
- **Token blocklist**: Currently in-memory + DB. For multi-instance, use a shared DB or Redis
- **Database**: Single PostgreSQL instance. Consider read replicas for read-heavy workloads
- **File uploads**: Not included. Add S3/GCS if needed
