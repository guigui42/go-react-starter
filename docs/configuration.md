# Configuration Reference

All configuration is via environment variables. Set them in `backend/.env` for development.

## Server

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `TRUST_PROXY` | `false` | Trust X-Forwarded-For headers |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

## Database

| Variable | Default | Description |
|----------|---------|-------------|
| `PGHOST` | `localhost` | PostgreSQL host |
| `PGPORT` | `5434` | PostgreSQL port |
| `PGUSER` | `postgres` | Database user |
| `PGPASSWORD` | — | Database password (**required**) |
| `PGDATABASE` | `starter` | Database name |
| `PGSSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |
| `PGSCHEMA` | `public` | Default schema |
| `AUTO_MIGRATE` | `true` | Run GORM AutoMigrate on startup |

## Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | — | JWT signing key (**required**, min 32 chars) |
| `ENCRYPTION_KEY` | — | Data encryption key (**required**) |
| `SESSION_DURATION` | `24h` | JWT token lifetime (Go duration) |

## WebAuthn / Passkeys

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBAUTHN_RP_ID` | `localhost` | Relying Party ID (your domain) |
| `WEBAUTHN_RP_ORIGIN` | `http://localhost:5173` | Allowed origin for WebAuthn |
| `WEBAUTHN_RP_NAME` | `Go React Starter` | Display name in passkey prompts |
| `WEBAUTHN_TIMEOUT` | `60000` | WebAuthn timeout in milliseconds |

## Admin

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_EMAILS` | — | Comma-separated list of admin email addresses |

Users who register with these emails are automatically promoted to admin.

## OAuth Providers

### GitHub

| Variable | Default | Description |
|----------|---------|-------------|
| `OAUTH_GITHUB_CLIENT_ID` | — | GitHub OAuth App client ID |
| `OAUTH_GITHUB_CLIENT_SECRET` | — | GitHub OAuth App client secret |

### Google

| Variable | Default | Description |
|----------|---------|-------------|
| `OAUTH_GOOGLE_CLIENT_ID` | — | Google OAuth client ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | — | Google OAuth client secret |

### Facebook

| Variable | Default | Description |
|----------|---------|-------------|
| `OAUTH_FACEBOOK_CLIENT_ID` | — | Facebook App ID |
| `OAUTH_FACEBOOK_CLIENT_SECRET` | — | Facebook App secret |

### OAuth General

| Variable | Default | Description |
|----------|---------|-------------|
| `OAUTH_CALLBACK_URL` | `http://localhost:8080` | Base URL for OAuth callbacks |
| `OAUTH_STATE_DURATION` | `10m` | OAuth state parameter lifetime |

## Email (SMTP)

| Variable | Default | Description |
|----------|---------|-------------|
| `SMTP_HOST` | — | SMTP server hostname |
| `SMTP_PORT` | `587` | SMTP server port |
| `SMTP_USERNAME` | — | SMTP username |
| `SMTP_PASSWORD` | — | SMTP password |
| `SMTP_USE_TLS` | `true` | Use TLS for SMTP |
| `EMAIL_FROM` | — | Sender email address |
| `EMAIL_FROM_NAME` | `Go React Starter` | Sender display name |
| `EMAIL_VERIFY_REQUIRED` | `false` | Require email verification before login |
| `EMAIL_TEMPLATES_PATH` | `./templates/emails` | Path to email templates |
| `EMAIL_BASE_URL` | `http://localhost:5173` | Base URL for email links |

If SMTP is not configured, a no-op email provider is used (emails are logged but not sent).

## Observability

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENABLED` | `false` | Enable OpenTelemetry |
| `OTEL_SERVICE_NAME` | `go-react-starter-api` | Service name for traces |
| `PROMETHEUS_ENABLED` | `false` | Enable Prometheus metrics endpoint |
| `OTEL_TRACING_ENABLED` | `false` | Enable distributed tracing |
| `OTEL_LOGS_ENABLED` | `false` | Enable log export via OTel |

## Docker Compose

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_PORT` | `5434` | Host port for PostgreSQL container |
