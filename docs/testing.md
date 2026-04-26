# Testing

## Overview

The project uses:
- **Backend**: Go's `testing` package with `testify` assertions
- **Frontend**: Vitest with React Testing Library

## Backend Testing

### Running Tests

```bash
make test-backend       # Run all Go tests
cd backend && make test # Same thing
```

### Test Database Isolation

Each test gets its own PostgreSQL schema, enabling `t.Parallel()` safely:

```
┌──────────────────────────────────────────────┐
│ PostgreSQL (testcontainers)                  │
│                                              │
│  template_schema (created once)              │
│  ├── All tables from AllModels()             │
│  │                                           │
│  test_abc123 (cloned per test)               │
│  ├── Independent data                        │
│  │                                           │
│  test_def456 (cloned per test)               │
│  ├── Independent data                        │
└──────────────────────────────────────────────┘
```

### How It Works

1. **First test**: Starts a PostgreSQL container via testcontainers-go
2. **Template schema**: Created once, runs `AutoMigrate(AllModels()...)`
3. **Per-test schema**: Each test calls `testutil.SetupTestDB(t)` which:
   - Creates a unique schema (`test_<uuid>`)
   - Clones the template schema's tables
   - Sets `search_path` in the GORM DSN
   - Registers `t.Cleanup()` to drop the schema after the test
4. **Parallel safe**: Each test operates on its own schema

### Using the Test Helper

```go
func TestCreateNote(t *testing.T) {
    t.Parallel()
    db := testutil.SetupTestDB(t)

    repo := repository.NewNoteRepository(db)
    note := &models.Note{
        UserID: uuid.New(),
        Title:  "Test note",
    }

    err := repo.Create(context.Background(), note)
    require.NoError(t, err)
    assert.NotEmpty(t, note.ID)
}
```

### External Database

Skip testcontainers by setting `TEST_DB_URL`:

```bash
TEST_DB_URL="postgres://user:pass@localhost:5432/testdb" make test-backend
```

### Test Patterns

**Handler tests** — Test HTTP handlers with `httptest`:

```go
func TestListNotes(t *testing.T) {
    db := testutil.SetupTestDB(t)
    handler := handlers.NewNoteHandler(repository.NewNoteRepository(db))

    req := httptest.NewRequest("GET", "/api/v1/notes", nil)
    // Add auth context...
    w := httptest.NewRecorder()

    handler.List(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

**Repository tests** — Test data access directly:

```go
func TestNoteRepository_ListByUser(t *testing.T) {
    db := testutil.SetupTestDB(t)
    repo := repository.NewNoteRepository(db)

    userID := uuid.New()
    // Create test data...

    notes, err := repo.ListByUser(context.Background(), userID)
    require.NoError(t, err)
    assert.Len(t, notes, expectedCount)
}
```

## Frontend Testing

### Running Tests

```bash
make test-frontend          # Run all Vitest tests
cd frontend && npm run test # Same thing
```

### Test Structure

Tests are co-located with components:

```
features/notes/
├── components/
│   ├── NoteCard.tsx
│   └── NoteCard.test.tsx    # Unit test next to component
├── hooks/
│   ├── index.ts
│   └── index.test.ts        # Hook tests
```

### Testing Patterns

**Component tests** with React Testing Library:

```tsx
import { render, screen } from '@testing-library/react'
import { NoteCard } from './NoteCard'

test('renders note title', () => {
  render(<NoteCard note={{ id: '1', title: 'Hello', content: '' }} />)
  expect(screen.getByText('Hello')).toBeInTheDocument()
})
```

**Hook tests** with `renderHook`:

```tsx
import { renderHook, waitFor } from '@testing-library/react'
import { useNotes } from './hooks'

test('fetches notes', async () => {
  const { result } = renderHook(() => useNotes(), { wrapper: QueryProvider })
  await waitFor(() => expect(result.current.isSuccess).toBe(true))
  expect(result.current.data).toBeDefined()
})
```

## CI/CD Test Pipelines

### Backend (`backend-tests.yml`)

- Runs on Ubuntu with PostgreSQL 17 service container
- Sets `TEST_DB_URL` to the CI PostgreSQL instance
- Executes: `go test -race -v ./...`
- Also runs `go vet` and linting

### Frontend (`frontend-tests.yml`)

- Runs on Ubuntu with Node.js 25
- Executes: `npm run test -- --run` (single run, no watch)
- Also runs: `npx tsc --noEmit` (type checking)
- Also runs: `npm run lint`

## Best Practices

1. **Use `t.Parallel()`** — Tests run faster and catch race conditions
2. **Use `testutil.SetupTestDB(t)`** — Automatic cleanup, no shared state
3. **Test at the right level** — Unit test business logic, integration test handlers
4. **Use `require` for setup** — `require.NoError(t, err)` stops the test on failure
5. **Use `assert` for checks** — `assert.Equal(t, ...)` continues on failure to report all issues
6. **Don't mock the database** — Real PostgreSQL via testcontainers is fast and accurate
