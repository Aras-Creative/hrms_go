# Module Roadmap

Current state of all domain modules, gaps, and next steps.

---

## Module Status

| Module | Rating | Entity Logic | Usecase Interfaces | Adapters | Docs | Tests |
|--------|--------|-------------|-------------------|----------|------|-------|
| `contract` | Polished | Yes (10 methods) | Yes (13) | Yes (5) | Yes | No |
| `leave` | Polished | Yes (11 methods) | Yes (3) | Yes (4) | Yes | No |
| `attendance` | Functional | Yes (11 methods) | Yes (2) | Yes (2) | Yes | No |
| `auth` | Functional | Yes (13 methods) | Yes (3) | Yes (4) | No | No |
| `payroll` | Functional | Thin (5 methods) | Yes (1) | Yes (1) | Yes | No |
| `employee` | Functional | Yes (15 methods) | Yes (4) | Yes (4) | No | No |
| `schedule` | Functional | Yes (4 methods) | Yes (1) | Yes (1) | No | No |
| `designation` | Rough | Thin (1 method) | No | No | No | No |
| `notification` | Rough | None | No | No | No | No |
| `storage` | Rough | None | No | No | No | No |
| `events` | Minimal | None | None | No | No | No |
| `audit` | Minimal | Thin (1 method) | No | No | No | No |

---

## Gaps

### Critical

- **Zero tests across all 15 modules.** No `_test.go` files exist anywhere under `internal/`.
- Start with `contract` and `leave` — richest entity logic, highest value.

### High

- **9/11 domain modules missing docs.** Only `contract` and `payroll` have docs in `docs/`.
- **3 modules lack usecase interfaces:** `designation`, `notification`, `audit`. Direct repo dependencies make testing and cross-module changes harder.

### Medium

- **4 modules lack adapter layer:** `designation`, `notification`, `storage`, `events`. No dependency inversion.
- **`notification` has zero entity methods.** Pure struct + pass-through usecase. No domain logic.
- **`storage` has no `routes.go`.** Handler exists but route registration is wired differently.
- **`audit` has no HTTP handler.** Used internally only — acceptable if intentional, but limits admin visibility.

### Low

- **`events` is delivery-only.** 2 files (handler + routes). SSE multiplexer with no domain. Acceptable if it stays thin.
- **`storage/repository/r2.go`** has the only `context.TODO()` in the codebase.
- **`payroll` usecase takes 9+ repo dependencies.** Consider splitting into sub-usecases.

---

## Priority Order

### Phase 1 — Polish (closest to done)

1. **`leave`** — ✅ Docs added (`docs/leave-module.md`). Entity has `Approve`/`Reject`/`Cancel` state machine + `LeaveBalance` quota enforcement.

### Phase 2 — Functional modules (need docs + possible interface extraction)

2. **`attendance`** — ✅ Docs added (`docs/attendance-module.md`). Rich entity (11 methods), has adapters, scheduler, corrections, SSE.
3. **`auth`** — Rich entity across 5 files. Has adapters. Needs docs.
4. **`employee`** — Rich value objects. Has adapters + `numbergen/`. Needs docs.
5. **`schedule`** — Clean but thin. Needs docs.

### Phase 3 — Rough modules (need structural work)

6. **`designation`** — Add usecase interfaces, adapters, models/ separation.
7. **`notification`** — Add entity methods, usecase interfaces, adapters.
8. **`storage`** — Add `routes.go`, usecase interfaces, adapters.

### Phase 4 — Minimal modules (decide scope)

9. **`audit`** — Decide: internal-only or add HTTP handler for admin visibility?
10. **`events`** — Decide: keep as thin transport or add domain logic?

### Ongoing

- **Write tests** starting from Phase 1. Entity logic tests first (state machines, guards, value objects), then usecase tests with mocked interfaces.
