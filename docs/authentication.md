# Authentication

Go React Starter ships with a complete, production-grade authentication system supporting multiple methods.

## Auth Methods

### Email/Password

Standard registration and login with:
- **bcrypt** password hashing (cost factor 12)
- Account **lockout** after repeated failures (configurable)
- Optional **email verification** before login

```
POST /auth/register  { email, password }
POST /auth/login     { email, password }
POST /auth/logout
GET  /auth/me        → current user (requires auth)
```

### Passkeys (WebAuthn)

Passwordless authentication using FIDO2/WebAuthn:
- Register passkeys from the settings page
- Login with biometrics, security keys, or platform authenticators
- Manage (rename, delete) passkeys

```
POST /auth/passkey/authenticate/begin   → challenge
POST /auth/passkey/authenticate/finish  → JWT
POST /auth/passkey/register/begin       → challenge (auth required)
POST /auth/passkey/register/finish      → credential (auth required)
```

**Configuration** (`.env`):

```bash
WEBAUTHN_RP_ID=localhost                    # Relying party ID (your domain)
WEBAUTHN_RP_ORIGIN=http://localhost:5173    # Allowed origin
WEBAUTHN_RP_NAME=My App                     # Display name
```

### OAuth (Google, GitHub, Facebook)

Social login with automatic account linking:
- First login creates account
- Subsequent logins link to existing account (by email)
- Users can link/unlink providers from settings

```
GET  /auth/oauth/providers           → enabled providers list
GET  /auth/oauth/{provider}          → redirect to provider
GET  /auth/oauth/{provider}/callback → handle callback, set JWT
```

**Configuration** (`.env`):

```bash
OAUTH_GITHUB_CLIENT_ID=...
OAUTH_GITHUB_CLIENT_SECRET=...
OAUTH_GOOGLE_CLIENT_ID=...
OAUTH_GOOGLE_CLIENT_SECRET=...
OAUTH_CALLBACK_URL=http://localhost:8080
```

### Backup Codes

One-time recovery codes for when primary auth is unavailable:
- Generated as a set of 10 codes
- Each code can only be used once
- Accessible from the passkey migration flow

```
POST /auth/migration/generate-backup-codes → 10 codes
POST /auth/backup-code/authenticate        → JWT
```

## JWT Session Management

### Token Flow

1. User authenticates → backend sets `auth_token` httpOnly cookie
2. Frontend sends cookie automatically with `credentials: 'include'`
3. Auth middleware validates JWT on every protected request
4. Logout adds token JTI to blocklist

### Cookie Configuration

| Property | Value | Purpose |
|----------|-------|---------|
| `HttpOnly` | `true` | Prevents XSS access |
| `Secure` | `true` (prod) | HTTPS only |
| `SameSite` | `Lax` | CSRF protection layer |
| `Path` | `/` | Available to all routes |
| `Max-Age` | `SESSION_DURATION` | Default 24h |

### CSRF Protection

Double-submit cookie pattern:
1. Backend sets `csrf_token` cookie (readable by JS)
2. Frontend reads cookie and sends `X-CSRF-Token` header
3. Backend verifies header matches cookie
4. Applied to all state-changing requests (POST, PUT, DELETE)

### Token Blocklist

On logout, the JWT's JTI (unique ID) is added to a blocklist:
- **In-memory cache** for fast lookups
- **Database persistence** for crash recovery
- Expired entries are automatically cleaned up

## Frontend Auth Integration

### Auth Context

```tsx
import { useAuth } from '@/features/auth/hooks'

function MyComponent() {
  const { user, isAuthenticated, isLoading, logout } = useAuth()

  if (isLoading) return <Spinner />
  if (!isAuthenticated) return <Navigate to="/login" />

  return <div>Welcome, {user.email}</div>
}
```

### Protected Routes

Routes requiring authentication redirect to `/login` automatically. Admin routes also check the `is_admin` flag.

### Multi-Method Login

The login page uses `POST /auth/check-methods` to detect available auth methods for an email (password, passkey, OAuth providers) and shows the appropriate UI.
