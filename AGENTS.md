# AGENTS.md

Architecture conventions and coding rules for this Go HRMS codebase.

---

## Module Layout

Every business domain lives under `internal/<module>/` with this structure:

```
internal/<module>/
├── entity/           Domain entities, value objects, status enums
├── models/           Cross-layer DTOs (input/output structs for usecases)
├── repository/       Interface definitions + Postgres implementation + DB models
├── usecase/          Business logic + port interfaces for cross-module deps
├── adapter/          Bridges this module's ports to other modules' repos/usecases
└── delivery/         HTTP handler, routes, request/response DTOs
```

Optional: `numbergen/` (sequential IDs), `processor/` (background logic), `templates/` (embedded files).

Non-module packages: `internal/pkg/` (shared infra), `internal/bootstrap/` (wiring), `internal/server/` (Fiber app).

---

## Entity Conventions

### Constructors

Every entity has two constructors:

```go
// NewXxx — assigns uuid.New(), sets time.Now(), applies business defaults
func NewXxx(...) *Xxx

// ReconstituteXxx — accepts ALL fields, no business logic. Used by repository only.
func ReconstituteXxx(...) *Xxx
```

### Value Objects

Structs with private `value` field, validated constructor, and DB bypass:

```go
type Phone struct { value string }
func NewPhone(value string) (Phone, error)     // validates
func PhoneFromDB(value string) Phone            // bypass (repo layer)
func (p Phone) String() string
```

### Status Enums

```go
type LeaveStatus string
const (
    LeaveStatusPending  LeaveStatus = "pending"
    LeaveStatusApproved LeaveStatus = "approved"
)
var validLeaveStatuses = []LeaveStatus{...}
func ParseLeaveStatus(s string) (LeaveStatus, error)
func (s LeaveStatus) IsValid() bool
```

### State Machines

Enforce transitions via guarded methods on entities. Never mutate status directly in usecases:

```go
func (c *Contract) MarkSent() error {
    if c.Status != ContractStatusDraft {
        return fmt.Errorf("contract must be in draft status to mark as sent")
    }
    c.Status = ContractStatusSent
    c.UpdatedAt = time.Now()
    return nil
}
```

Or via transition map:

```go
var allowedTransitions = map[LeaveStatus][]LeaveStatus{...}
func (s LeaveStatus) CanTransitionTo(next LeaveStatus) bool { ... }
```

---

## Repository Conventions

### Interface Definition (`repository.go`)

Interfaces only, no implementation. One file per aggregate:

```go
type EmployeeRepository interface {
    Create(ctx context.Context, e *entity.Employee) error
    FindByID(ctx context.Context, id string) (*entity.Employee, error)
    Update(ctx context.Context, e *entity.Employee) error
    Delete(ctx context.Context, id string) error
}
```

### Implementation (`postgres.go`)

- Struct: `Postgres<Type>Repo`
- Constructor: `NewPostgres<Type>Repo(db *sqlx.DB)`
- Raw SQL as `const` at top of file, composed via string concatenation
- `sql.ErrNoRows` returns `nil, nil` — not-found is not an error

### DB Models (`models.go`)

Separate from domain entities. `db:"column_name"` tags only:

```go
type EmployeeModel struct {
    ID             string  `db:"id"`
    PersonalEmail  *string `db:"personal_email"`
}
```

### Model-to-Entity

Private `modelToEntity()` calls `entity.ReconstituteXxx()` with value object bypass constructors:

```go
func modelToEntity(m *EmployeeModel) *entity.Employee {
    return entity.ReconstituteEmployee(
        m.ID, m.UserID, m.FullName,
        entity.PhoneFromDB(m.Phone),
        entity.Gender(m.Gender),
        ...
    )
}
```

---

## Usecase Conventions

### Struct & Constructor

Concrete struct named `<Domain>Usecase`. Receives interfaces for cross-module deps, concrete types for same-module repos:

```go
type ContractUsecase struct {
    tmplRepo    repository.TemplateRepository   // same module
    empFetcher  EmployeeFetcher                  // cross-module (interface)
}

func NewContractUsecase(tmplRepo repository.TemplateRepository, empFetcher EmployeeFetcher) *ContractUsecase {
    return &ContractUsecase{tmplRepo: tmplRepo, empFetcher: empFetcher}
}
```

### Port Interfaces

Define **local interfaces** in the usecase file for each cross-module capability needed. Named by capability, not source module:

```go
type EmployeeFetcher interface {
    FindByID(ctx context.Context, id string) (*empEntity.Employee, error)
    FindDesignationIDs(ctx context.Context, ids []string) (map[string]*string, error)
}

type UserDeactivator interface {
    Deactivate(ctx context.Context, userID string) error
}
```

### Error Handling

Import as `errors "hrms/internal/pkg/apperror"`. Use `DomainError` for client-facing errors:

```go
// Not-found → 404
return nil, errors.NewNotFound("employee not found")

// Validation → 400
return nil, errors.NewInvalidInput("invalid start_date")

// Business rule / internal → wrapped fmt.Errorf
return nil, fmt.Errorf("find contract: %w", err)
```

---

## Adapter Conventions

One adapter per port interface. Lives in the consuming module's `adapter/`:

```go
package adapter

import (
    emplRepo "hrms/internal/employee/repository"
    contractUc "hrms/internal/contract/usecase"
)

type EmployeeFetcherAdapter struct {
    repo emplRepo.EmployeeRepository
}

func NewEmployeeFetcherAdapter(repo emplRepo.EmployeeRepository) *EmployeeFetcherAdapter {
    return &EmployeeFetcherAdapter{repo: repo}
}

func (a *EmployeeFetcherAdapter) FindByID(ctx context.Context, id string) (*emplEntity.Employee, error) {
    return a.repo.FindByID(ctx, id)
}

var _ contractUc.EmployeeFetcher = (*EmployeeFetcherAdapter)(nil)
```

Every adapter file ends with a `var _` compile-time check.

---

## Delivery Conventions

### File Layout

- `handler.go` — Handler struct + methods
- `routes.go` — `RegisterRoutes(r fiber.Router, authMw fiber.Handler)`
- `request.go` — Request DTOs with `json:` + `validate:` tags
- `response.go` — Response DTOs with `json:` tags + `toXxxResponse()` converters

### Handler Method Pattern

```go
func (h *EmployeeHandler) Create(c fiber.Ctx) error {
    var req CreateRequest
    if err := c.Bind().Body(&req); err != nil {
        return err
    }

    input := models.CreateEmployeeInput{...}   // map request → usecase input
    e, err := h.uc.Create(c.RequestCtx(), input)
    if err != nil {
        return response.Error(c, err)           // centralized error mapping
    }
    return response.Created(c, toResponse(e))   // entity → response conversion
}
```

### Response Helpers

- `response.OK(c, data)` — 200
- `response.Created(c, data)` — 201
- `response.Paginate(c, items, total, page, perPage)` — 200 with pagination
- `response.NoContent(c)` — 204
- `response.Error(c, err)` — maps `DomainError` to HTTP status, unknown errors → 500

### Routes

Middleware injected per-route or per-group:

```go
func (h *ContractHandler) RegisterRoutes(r fiber.Router, authMw fiber.Handler) {
    r.Get("/contracts", authMw, h.ListContracts)
    r.Post("/contracts", authMw, h.CreateContract)
    r.Get("/contracts/:id/preview", authMw, h.PreviewContract)
}
```

All routes must use `authMw` unless explicitly public.

---

## Bootstrap Wiring

Single file `internal/bootstrap/bootstrap.go`, function `Run(cfgPath string)`. Wire in dependency order:

```go
// 1. Repos
emplRepo := emplRepo.NewPostgresEmployeeRepo(db)

// 2. Adapters (bridge to other modules)
designationFetcher := emplAdapter.NewDesignationFetcherAdapter(designationRepo)

// 3. Usecase (inject repo + adapters)
emplUC := emplUC.NewEmployeeUsecase(emplRepo, designationFetcher, ...)

// 4. Handler
emplHandler := emplDelivery.NewEmployeeHandler(emplUC)

// 5. Routes
emplHandler.RegisterRoutes(api, authmw)
```

Add `var _` compile-time checks when repos are passed directly as port interfaces:

```go
var _ contractUc.UserDeactivator = userRepo
var _ contractUc.SessionRevoker = sessionRepo
```

---

## Import Aliases

Pattern: `<module><layer>`

| Layer | Alias |
|-------|-------|
| repository | `emplRepo`, `contractRepo`, `leaveRepo` |
| usecase | `emplUc`, `contractUc`, `leaveUc` |
| adapter | `emplAdapter`, `contractadapter`, `leaveAdapter` |
| delivery | `emplDelivery`, `contractDelivery`, `leaveDelivery` |
| entity | `emplEntity`, `contractEntity`, `empEntity` |

Shared packages:
- `errors "hrms/internal/pkg/apperror"` — domain errors
- `response "hrms/internal/pkg/api"` — HTTP response helpers

---

## Migration Convention

File: `NNNNNN_<snake_case_description>.{up,down}.sql`

- 6-digit zero-padded sequential number
- Descriptions: `create_<table>_table`, `add_<column>`, `drop_<table>_column`
- DB conventions: UUID PKs (`DEFAULT gen_random_uuid()`), TIMESTAMPTZ, plural table names, snake_case columns

---

## Compile-Time Checks

Every adapter file:
```go
var _ <package>.PortInterface = (*AdapterStruct)(nil)
```

Bootstrap (when repos satisfy ports directly):
```go
var _ <package>.PortInterface = concreteRepo
```

---

## Error Response Shape

```json
{
    "success": false,
    "error": {
        "code": "NOT_FOUND",
        "message": "employee not found"
    }
}
```

Success:
```json
{
    "success": true,
    "data": { ... }
}
```

Paginated:
```json
{
    "success": true,
    "data": [...],
    "pagination": {
        "page": 1,
        "per_page": 20,
        "total": 42,
        "total_pages": 3
    }
}
```
