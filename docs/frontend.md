# Frontend Guide

## Tech Stack

| Library | Version | Purpose |
|---------|---------|---------|
| React | 19+ | UI framework |
| TypeScript | 5+ | Type safety |
| Vite | 8+ | Build tool + HMR |
| TanStack Router | 1+ | File-based routing |
| TanStack Query | 5+ | Server state management |
| Mantine | 9+ | UI component library |
| i18next | — | Internationalization (EN/FR) |

## Project Structure

```
frontend/src/
├── components/          # Shared components (Layout, ErrorBoundary, etc.)
├── features/            # Feature modules
│   ├── auth/           # Authentication (login, register, OAuth, passkeys)
│   ├── admin/          # Admin dashboard
│   ├── notes/          # Sample CRUD feature
│   └── user/           # User settings
├── i18n/               # Translation files
│   └── locales/
│       ├── en.json     # English translations
│       └── fr.json     # French translations
├── lib/                # Utilities
│   ├── api.ts          # API client (apiRequest, apiDownload)
│   └── queryClient.ts  # TanStack Query configuration
├── routes/             # File-based routes
│   ├── __root.tsx      # Root layout
│   ├── index.tsx       # Home page
│   ├── login.tsx       # Login page
│   ├── register.tsx    # Registration
│   ├── settings.tsx    # User settings
│   ├── notes/          # Notes pages
│   └── admin/          # Admin pages
└── theme/              # Mantine theme configuration
```

## Routing

TanStack Router with file-based routing. Routes map to files in `src/routes/`:

```
/              → routes/index.tsx
/login         → routes/login.tsx
/register      → routes/register.tsx
/settings      → routes/settings.tsx
/notes         → routes/notes/index.tsx
/notes/:noteId → routes/notes/$noteId.tsx
/admin         → routes/admin/index.tsx
```

### Adding a Route

Create a file in `src/routes/`:

```tsx
// src/routes/tasks/index.tsx
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/tasks/')({
  component: TasksPage,
})

function TasksPage() {
  return <div>Tasks</div>
}
```

TanStack Router auto-generates the route tree — no manual registration needed.

## API Client

`src/lib/api.ts` provides `apiRequest<T>()`:

```typescript
import { apiRequest } from '@/lib/api'

// GET request
const notes = await apiRequest<Note[]>('/api/v1/notes')

// POST request
const note = await apiRequest<Note>('/api/v1/notes', {
  method: 'POST',
  body: JSON.stringify({ title: 'New note', content: 'Hello' }),
})
```

Features:
- Automatic `Content-Type: application/json`
- CSRF token injection for state-changing requests
- Cookie-based auth (`credentials: 'include'`)
- Backend `{ data: ... }` wrapper unwrapping
- Typed error handling via `ApiError` class

## State Management

### Server State (TanStack Query)

All server data flows through TanStack Query:

```typescript
// Fetch data
const { data, isLoading, error } = useQuery({
  queryKey: ['notes'],
  queryFn: () => apiRequest<Note[]>('/api/v1/notes'),
})

// Mutate data
const mutation = useMutation({
  mutationFn: (data: CreateNoteRequest) =>
    apiRequest<Note>('/api/v1/notes', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['notes'] })
  },
})
```

### Auth State

Auth state is managed via a React context that checks `GET /auth/me`:

```typescript
const { user, isAuthenticated, isLoading, logout } = useAuth()
```

## Internationalization (i18n)

Supports English and French out of the box.

### Translation Files

```
src/i18n/locales/
├── en.json    # English
└── fr.json    # French
```

### Using Translations

```tsx
import { useTranslation } from 'react-i18next'

function MyComponent() {
  const { t } = useTranslation()
  return <h1>{t('common.welcome')}</h1>
}
```

### Adding a Language

1. Create `src/i18n/locales/de.json` (copy structure from `en.json`)
2. Register in `src/i18n/index.ts`
3. Add language option to the settings page

### Translation Key Conventions

```json
{
  "featureName": {
    "pageTitle": "...",
    "buttons": {
      "save": "Save",
      "cancel": "Cancel"
    },
    "messages": {
      "success": "...",
      "error": "..."
    }
  }
}
```

## Mantine UI

### Theme

Custom theme in `src/theme/index.ts`:
- Custom color palette (AppDark, AppBlue)
- Dark mode by default
- CSS variable resolver for custom properties
- WCAG 2.1 AA accessible by default

### Using Components

```tsx
import { Button, TextInput, Stack, Paper } from '@mantine/core'

function MyForm() {
  return (
    <Paper p="md" withBorder>
      <Stack gap="sm">
        <TextInput label="Title" required />
        <Button type="submit">Save</Button>
      </Stack>
    </Paper>
  )
}
```

## PWA Support

Progressive Web App is pre-configured:
- Service worker via `vite-plugin-pwa`
- Offline fallback page
- App manifest with icons
- Install prompt handling

## Development

```bash
npm run dev          # Start Vite dev server (port 5173)
npm run build        # Production build
npm run preview      # Preview production build
npm run test         # Run Vitest
npm run lint         # ESLint
npm run format       # Prettier
npm run type-check   # TypeScript compiler check
```
