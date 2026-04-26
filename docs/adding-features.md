# Adding Features

This guide walks through adding a new feature end-to-end, using the included **Notes** feature as a reference.

## Overview

A full-stack feature requires:

1. **Model** — GORM struct (database table)
2. **Migration** — Schema change
3. **Repository** — Data access with user scoping
4. **Handler** — HTTP endpoints
5. **Route registration** — Wire into the router
6. **Frontend types** — TypeScript interfaces
7. **Frontend hooks** — TanStack Query hooks
8. **Frontend components** — React UI
9. **Frontend route** — Page in the router

## Step 1: Backend Model

Create `backend/internal/models/your_feature.go`:

```go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Task struct {
    ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
    Title     string         `gorm:"not null;size:255" json:"title"`
    Done      bool           `gorm:"default:false" json:"done"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

**Important**: Include `UserID` with an index for row-level security.

### Register the Model

Add to `backend/internal/models/registry.go`:

```go
func AllModels() []interface{} {
    return []interface{}{
        // ... existing models ...
        &Task{},
    }
}
```

### Register as User-Scoped Table

Add `"tasks"` to the protected tables list in `backend/internal/repository/scopes/user_scope_guard.go`:

```go
var userScopedTables = map[string]bool{
    "notes": true,
    "tasks": true,  // ← add here
}
```

This ensures every query on `tasks` requires a `user_id` filter (panics in dev if missing).

## Step 2: Migration

Create `backend/internal/migrations/002_add_tasks.go`:

```go
package migrations

import "gorm.io/gorm"

func init() {
    Register(Migration{
        Version: "002",
        Name:    "add_tasks",
        Up: func(db *gorm.DB) error {
            type Task struct {
                ID     string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
                UserID string `gorm:"type:uuid;not null;index"`
                Title  string `gorm:"not null;size:255"`
                Done   bool   `gorm:"default:false"`
            }
            return db.AutoMigrate(&Task{})
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropTable("tasks")
        },
    })
}
```

**Convention**: Version is zero-padded (`"002"`, `"003"`, ...). Name is `snake_case`.

## Step 3: Repository

Create `backend/internal/repository/task_repository.go`:

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/example/go-react-starter/internal/models"
    "github.com/example/go-react-starter/internal/repository/scopes"
    "gorm.io/gorm"
)

type TaskRepository struct {
    db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
    return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
    return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Task, error) {
    var tasks []models.Task
    err := r.db.WithContext(ctx).
        Scopes(scopes.ForUser(userID)).
        Order("created_at DESC").
        Find(&tasks).Error
    return tasks, err
}

func (r *TaskRepository) GetByID(ctx context.Context, userID, taskID uuid.UUID) (*models.Task, error) {
    var task models.Task
    err := r.db.WithContext(ctx).
        Scopes(scopes.ForUser(userID)).
        Where("id = ?", taskID).
        First(&task).Error
    return &task, err
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
    return r.db.WithContext(ctx).
        Scopes(scopes.ForUser(task.UserID)).
        Save(task).Error
}

func (r *TaskRepository) Delete(ctx context.Context, userID, taskID uuid.UUID) error {
    return r.db.WithContext(ctx).
        Scopes(scopes.ForUser(userID)).
        Where("id = ?", taskID).
        Delete(&models.Task{}).Error
}
```

**Key pattern**: Every query uses `scopes.ForUser(userID)` for row-level security.

## Step 4: Handler

Create `backend/internal/handlers/task.go`:

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/example/go-react-starter/internal/middleware"
    "github.com/example/go-react-starter/internal/models"
    "github.com/example/go-react-starter/internal/repository"
    "github.com/example/go-react-starter/pkg/response"
)

type TaskHandler struct {
    repo *repository.TaskRepository
}

func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
    return &TaskHandler{repo: repo}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := middleware.UserIDFromContext(r.Context())

    var req struct {
        Title string `json:"title"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "bad_request", "Invalid JSON", nil)
        return
    }

    task := &models.Task{
        UserID: userID,
        Title:  req.Title,
    }
    if err := h.repo.Create(r.Context(), task); err != nil {
        response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to create task", nil)
        return
    }
    response.Success(w, http.StatusCreated, task)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
    userID := middleware.UserIDFromContext(r.Context())
    tasks, err := h.repo.ListByUser(r.Context(), userID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "internal_error", "Failed to list tasks", nil)
        return
    }
    response.Success(w, http.StatusOK, tasks)
}

// ... GetByID, Update, Delete follow the same pattern
```

## Step 5: Register Routes

In `backend/cmd/server/main.go`, add inside the authenticated API group:

```go
taskRepo := repository.NewTaskRepository(db)
taskHandler := handlers.NewTaskHandler(taskRepo)

r.Route("/tasks", func(r chi.Router) {
    r.Post("/", taskHandler.Create)
    r.Get("/", taskHandler.List)
    r.Get("/{id}", taskHandler.GetByID)
    r.Put("/{id}", taskHandler.Update)
    r.Delete("/{id}", taskHandler.Delete)
})
```

## Step 6: Frontend Types

Create `frontend/src/features/tasks/types/index.ts`:

```typescript
export interface Task {
  id: string
  title: string
  done: boolean
  created_at: string
  updated_at: string
}

export interface CreateTaskRequest {
  title: string
}
```

## Step 7: Frontend Hooks

Create `frontend/src/features/tasks/hooks/index.ts`:

```typescript
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiRequest } from '@/lib/api'
import type { Task, CreateTaskRequest } from '../types'

const API_PREFIX = '/api/v1'

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: () => apiRequest<Task[]>(`${API_PREFIX}/tasks`),
  })
}

export function useCreateTask() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateTaskRequest) =>
      apiRequest<Task>(`${API_PREFIX}/tasks`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
  })
}
```

## Step 8: Frontend Route

Create `frontend/src/routes/tasks/index.tsx`:

```tsx
import { createFileRoute } from '@tanstack/react-router'
import { useTasks } from '@/features/tasks/hooks'

export const Route = createFileRoute('/tasks/')({
  component: TasksPage,
})

function TasksPage() {
  const { data: tasks, isLoading } = useTasks()
  // ... render tasks
}
```

## Step 9: Add Navigation

Add the route to `frontend/src/components/Layout.tsx` (or wherever your nav lives) and add translation keys to `frontend/src/i18n/locales/en.json` and `fr.json`.

## Checklist

- [ ] Model with `UserID` field and `gorm` tags
- [ ] Model added to `AllModels()` in `registry.go`
- [ ] Table added to `userScopedTables` in scope guard
- [ ] Migration file with `Up` and `Down`
- [ ] Repository with `scopes.ForUser()` on every query
- [ ] Handler using `response.Success/Error` patterns
- [ ] Routes registered in `main.go` under authenticated group
- [ ] Frontend types, hooks, components, route
- [ ] Translation keys for any user-facing text
- [ ] Admin dashboard updated if feature needs admin stats
