# Security

## Defense in Depth

The starter includes multiple overlapping security layers:

```
Request → Rate Limiter → CORS → Auth (JWT) → CSRF → Body Size → Handler
                                                                    ↓
                                                              Row-Level Security (GORM Scope)
                                                                    ↓
                                                              Audit Logging
```

## Middleware Reference

### Authentication (`middleware/auth.go`)

- Validates JWT from `auth_token` httpOnly cookie
- Injects user ID into request context
- Checks token against blocklist (revoked tokens)
- Two variants:
  - `AuthMiddleware` — returns 401 if not authenticated
  - `AuthMiddlewareWithUser` — also loads full user object into context

### CSRF Protection (`middleware/csrf.go`)

Double-submit cookie pattern:
1. Backend sets `csrf_token` cookie (JS-readable)
2. Frontend sends `X-CSRF-Token` header with same value
3. Backend verifies match on all POST/PUT/DELETE requests
4. GET/HEAD/OPTIONS are exempt

### Rate Limiting (`middleware/rate_limit.go`)

- **IP-based**: Limits requests per IP address
- **Per-user**: Limits requests per authenticated user
- Returns `429 Too Many Requests` with `Retry-After` header
- Configurable limits for auth endpoints (login, register)

### CORS (`middleware/cors.go`)

- Configured for development (localhost origins)
- In production, nginx handles CORS headers
- Only applied when `ENVIRONMENT != production`

### Body Size Limit (`middleware/body_size.go`)

- Enforces maximum request body size
- Returns `413 Payload Too Large` if exceeded
- Prevents memory exhaustion attacks

### Recovery (`middleware/recovery.go`)

- Catches panics in handlers
- Returns structured JSON error (not a stack trace)
- Logs full panic details to backend logs
- Prevents server crashes from becoming user-visible errors

### Token Blocklist (`middleware/token_blocklist.go`)

- On logout, JWT's JTI is added to blocklist
- In-memory cache for fast rejection
- Database persistence for crash recovery
- Automatic cleanup of expired entries

## Row-Level Security

GORM scopes enforce that users can only access their own data.

### How It Works

1. **ForUser scope**: Adds `WHERE user_id = ?` to every query
2. **UserScopeGuard callback**: GORM callback that validates queries at runtime
3. **Dev mode**: Guard **panics** if a user-scoped table is queried without `user_id`
4. **Production**: Guard logs a warning (defense in depth)

### Protected Tables

Any table listed in `userScopedTables` (in `scopes/user_scope_guard.go`) is protected. By default: `notes`, plus all auth-related tables.

### Usage Pattern

```go
// ✅ Correct — always scope by user
db.WithContext(ctx).Scopes(scopes.ForUser(userID)).Find(&notes)

// ✅ Admin bypass — for global queries in admin endpoints
adminCtx := scopes.SkipUserScopeGuard(ctx)
db.WithContext(adminCtx).Model(&models.Note{}).Count(&total)

// ❌ WILL PANIC IN DEV — no user_id filter
db.Find(&notes)
```

## Content Security Policy (CSP)

CSP headers are set in `docker/nginx.conf` for production:

```
default-src 'self';
script-src 'self' 'unsafe-inline';
style-src 'self' 'unsafe-inline';
frame-ancestors 'none';
base-uri 'self';
form-action 'self';
```

### CSP Violation Reporting

Violations are:
1. Sent to `POST /api/csp-report`
2. Logged via Zerolog
3. Persisted to `csp_violations` table
4. Viewable in admin dashboard (`GET /api/v1/admin/csp-violations`)

## Audit Logging

Sensitive operations are logged to the `audit_logs` table:

| Event Type | Actions |
|-----------|---------|
| `auth.login` | Success/failure, IP, user agent |
| `auth.register` | New account creation |
| `auth.logout` | Token invalidation |
| `auth.oauth` | OAuth authentication |
| `auth.passkey` | Passkey registration/authentication |
| `auth.password` | Password changes |
| `admin.*` | Admin actions |

Audit logs are queryable from the admin dashboard with filtering by event type, status, actor, and date range.

## Password Security

- **bcrypt** with cost factor 12 (tunable)
- Passwords never logged or returned in API responses
- Account lockout after configurable failed attempts
- Constant-time comparison to prevent timing attacks

## Best Practices for New Code

1. **Always use `scopes.ForUser()`** on queries for user-owned data
2. **Never log sensitive data** (passwords, tokens, keys)
3. **Validate all input** server-side (don't trust frontend validation)
4. **Use `response.Error()`** for error responses (consistent format)
5. **Add audit logging** for security-sensitive operations
6. **Register new user-scoped tables** in the scope guard
