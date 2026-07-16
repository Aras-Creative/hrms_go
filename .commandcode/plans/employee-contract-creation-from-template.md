# Create Employee Contract from Template with Overrides

## Overview

Add the ability to create employee contracts by hydrating from a template, applying overrides, and snapshotting the resolved data. The `contracts` table exists but has no Go CRUD — this implements the creation endpoint.

## New Files (7)

| File | Purpose |
|------|---------|
| `migrations/000023_add_contract_data_columns.up.sql` | Add `data` + `templates` JSONB columns to `contracts` |
| `migrations/000023_add_contract_data_columns.down.sql` | Rollback |
| `internal/contract/models/contract_create.go` | `CreateContractInput`, `ContractOverrides` |
| `internal/contract/delivery/contract_request.go` | `CreateContractRequest` + override DTOs + conversion |
| `internal/contract/delivery/contract_response.go` | `ContractResponse` + `toContractResponse()` |
| `internal/contract/usecase/contract.go` | `CreateContract()` usecase with merge logic |
| `internal/contract/adapter/employee_fetcher.go` | Adapter wrapping employee repo for usecase interface |

## Modified Files (5)

| File | Changes |
|------|---------|
| `internal/contract/entity/contract.go` | Add `Data ContractTemplateData`, `Templates ContractTemplatePartials` fields; update `NewContract()` / `ReconstituteContract()` |
| `internal/contract/repository/repository.go` | Add `CreateContract()` to interface, embed `numbergen.SequenceRepository` |
| `internal/contract/repository/postgres.go` | Add `ContractModel`, `CreateContract()` impl, 5 `SequenceRepository` methods |
| `internal/contract/delivery/handler.go` | Add `CreateContract()` handler method |
| `internal/contract/delivery/routes.go` | Add `POST /contracts` route |
| `internal/bootstrap/bootstrap.go` | Wire a `numbergen.Generator` for contracts, update `NewContractUsecase` call |

## Sequence Flow

```
POST /contracts
  Body: { template_id, employee_id, start_date, contract_date_legal, salary, end_date?, overrides? }

[1] Handler binds & validates dates (YYYY-MM-DD)
[2] Handler converts override DTOs → entity types
[3] Usecase loads template (404 if not found, 400 if inactive)
[4] Usecase looks up employee → gets DesignationID → gets designation Name
[5] Usecase copies template Data/Templates as defaults
[6] Usecase applies overrides via deepMerge (non-nil override fields win)
[7] Usecase generates contract number via numGen.Generate(ctx, "CTR")
[8] Usecase builds entity with status=draft
[9] Repository INSERT into contracts table
[10] Return 201 with full contract response
```

## Key Design Decisions

### Contract Number
- Reuse `numbergen.Generator` from `internal/employee/numbergen/`
- Prefix: `"2001"`, designation code: `"CTR"` (static)
- Produces: `2001CTR01`, `2001CTR02`, ...
- Single row in `number_sequences` table per contracts

### Designation Title Default
- No override → look up employee's `DesignationID`, then fetch designation `Name`
- Requires two adapters: `EmployeeFetcher` (returns `DesignationID`) and `DesignationFetcher` (returns `Name` by ID)
- Follows the existing adapter pattern (see `payroll/adapter/employee_fetcher.go`, `employee/adapter/designation_fetcher.go`)

### Override Merge Strategy
- Top-level fields (`designation_title`, `working_hours`): override if non-nil pointer
- `data`: field-by-field merge — non-nil/non-empty override fields replace template defaults
- `templates`: block-level replace — if override has non-nil `Blocks`, they fully replace
- Partial overrides supported: client can override only `JobDuties` without touching other fields

### PKWT Auto-End-Date
- If no `end_date` provided AND template type is PKWT → default to `start_date + 1 year`
- Explicit `end_date` always takes precedence

### Adapter Interfaces (defined in usecase package)

```go
type EmployeeFetcher interface {
    FindDesignationID(ctx context.Context, id string) (*string, error)
}

type DesignationFetcher interface {
    FindNameByID(ctx context.Context, id string) (string, error)
}
```

Both implemented as adapters in `internal/contract/adapter/` wrapping employee repo and designation repo.

## Database Migration

```sql
-- 000023_add_contract_data_columns.up.sql
ALTER TABLE contracts
    ADD COLUMN data      JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN templates JSONB NOT NULL DEFAULT '{}';

-- 000023_add_contract_data_columns.down.sql
ALTER TABLE contracts
    DROP COLUMN IF EXISTS data,
    DROP COLUMN IF EXISTS templates;
```

## Verification

1. Run `go build ./...` — no compile errors
2. Run existing tests — no regressions
3. Manual test: POST /contracts with valid template_id + employee_id + dates + salary
4. Manual test: POST /contracts with overrides on data fields, verify snapshot differs from template
5. Manual test: POST /contracts with inactive template → 400
6. Manual test: POST /contracts with nonexistent template → 404
7. Manual test: POST /contracts for PKWT without end_date → auto-calculated to +1 year
8. Manual test: POST /contracts without overrides → designation_title comes from employee's designation
9. Manual test: POST /contracts with explicit end_date → used as-is regardless of contract type
10. Verify contract number increments: `2001CTR01`, `2001CTR02`, ...
