# Payroll Module

Master data module for employee base salaries, compensations, benefits, and deductions.

---

## Module Structure

```
internal/payroll/
├── adapter/
│   └── employee_fetcher.go       # ExistsByID → employee repository
├── delivery/
│   ├── handler.go                # HTTP handlers (35 endpoints)
│   ├── request.go                # Request DTOs with validation tags
│   ├── response.go               # Response DTOs
│   └── routes.go                 # Route registration (admin-protected)
├── entity/
│   ├── value_objects.go          # Amount, Currency, Percentage, enums
│   ├── base_salary.go            # EmployeeBaseSalary
│   ├── compensation_item.go      # CompensationItem
│   ├── employee_compensation.go  # EmployeeCompensation (assignment)
│   ├── benefit_type.go           # BenefitType
│   ├── employee_benefit.go       # EmployeeBenefit (assignment)
│   ├── deduction_type.go         # DeductionType
│   └── employee_deduction.go     # EmployeeDeduction (assignment)
├── models/
│   └── payroll.go                # Usecase input DTOs
├── repository/
│   ├── repository.go             # 7 interfaces + filter structs
│   ├── models.go                 # 7 DB row structs (sqlx tags)
│   ├── postgres.go               # Salary + CompItem + EmpComp repos
│   ├── postgres_benefits.go      # BenefitType + EmployeeBenefit repos
│   └── postgres_deductions.go    # DeductionType + EmployeeDeduction repos
└── usecase/
    └── payroll.go                # PayrollUsecase — CRUD + EmployeeFetcher interface
```

---

## Routes

All under `/api/v1/payroll/`, admin-protected (authMw + adminMw).

| Path | CRUD |
|------|------|
| `/salary` | Base salaries |
| `/compensation` | Compensation items (master data) |
| `/compensation/assignments` | Per-employee compensation assignments |
| `/benefit` | Benefit types (master data) |
| `/benefit/assignments` | Per-employee benefit assignments |
| `/deduction` | Deduction types (master data) |
| `/deduction/assignments` | Per-employee deduction assignments |

Each group has standard REST: `GET /`, `GET /:id`, `POST /`, `PUT /:id`, `DELETE /:id`.

---

## Value Objects

| Type | Storage | Description |
|------|---------|-------------|
| `Amount` | `int64` cents | Monetary value, `NewAmount(float64)` / `AmountFromCents(int64)` |
| `Currency` | `string` (ISO 4217) | 3-letter code, validated by regex |
| `Percentage` | `float64` | 0.0–100.0 |
| `ContributionType` | enum | `percentage` / `fixed` |
| `DeductionCalcType` | enum | `percentage` / `fixed` |
| `Frequency` | enum | `monthly` / `yearly` / `one_time` |
| `CompensationItemType` | enum | `recurring` / `one_time` |

---

## Usecase Patterns

- **Create***: Validate value objects → check employee exists (via `EmployeeFetcher`) → build entity → save
- **Get***: FindByID → not found error
- **List***: Build filter → call repo findAll → return raw `([]*Entity, int64, error)` — pagination handled in delivery
- **Update***: FindByID → modify fields → save
- **Delete***: Hard delete from repo

---

## Response Mapping

All handler responses pass through `*ToResponse()` converters defined in `delivery/response.go` before serialization. This ensures:

- **Correct JSON tags**: snake_case field names via the `json:"..."` struct tags on response types
- **Value objects unwrapped**: e.g. `entity.Amount` (which wraps `int64 cents`) is output as `float64`, `entity.CompensationItemType` enum renders as a plain `string`
- **Consistent shape**: Same response struct used for single-item and list endpoints

Each entity has a pair of converters — singular and plural:

| Entity | Converters |
|--------|-----------|
| `EmployeeBaseSalary` | `baseSalaryToResponse()` / `baseSalariesToResponse()` |
| `CompensationItem` | `compItemToResponse()` / `compItemsToResponse()` |
| `EmployeeCompensation` | `empCompToResponse()` / `empCompsToResponse()` |
| `BenefitType` | `benefitTypeToResponse()` / `benefitTypesToResponse()` |
| `EmployeeBenefit` | `empBenefitToResponse()` / `empBenefitsToResponse()` |
| `DeductionType` | `deductionTypeToResponse()` / `deductionTypesToResponse()` |
| `EmployeeDeduction` | `empDeductionToResponse()` / `empDeductionsToResponse()` |

## Architecture Notes

- **Usecase returns raw data**: List methods return `([]*Entity, int64, error)` — no pagination wrapper at usecase level
- **Routes grouped by component**: Following `api-design` taste preference for sub-path grouping
- **Assignment endpoints**: Use `/assignments` suffix (not `employee-*` or `employee-assignments`)
- **Amounts in cents**: Salary and compensation stored as `BIGINT` (cents). Benefits and deductions use `DECIMAL(12,2)`
- **End date semantics**: `NULL` = currently active / indefinite
- **Employee validation**: Via `EmployeeFetcher` adapter at usecase layer
