# Attendance Architecture Rules

## Correction Always Wins

Admin corrections via `POST /attendance/corrections` are the highest-authority data in the attendance system. They cannot be overwritten by any automated process.

### Guards

| Layer | Guard |
|-------|-------|
| `ProcessDaily` | Returns early if `source == "correction"` — `GET /attendance/mine`, `GET /attendance/daily`, `GET /attendance/list` all see corrected values |
| `ProcessRange` (scheduler) | `WHERE NOT EXISTS (SELECT 1 FROM attendance_corrections ...)` in SQL — scheduler skips corrected employee+date combos entirely |
| `Delete correction` | Re-runs `ProcessDaily` to restore the original computed state (including re-marking `on_leave` if leave is still active) |

### Rationale
- The processor only runs **forward**. It processes current/future dates and never backfill-repairs past data.
- If past attendance data has errors, **admin must fix it manually** via a correction. Automated re-processing would overwrite the fix — the guards prevent that.
- There is no "re-process past dates" command. Past is owned by admin.

## Leave Immutability (Past Approved)

Approved leave with `end_date < today` cannot be cancelled or rejected.

| Action | Allowed? |
|--------|----------|
| Cancel pending leave (any date) | Yes |
| Cancel approved leave (`end_date >= today`) | Yes — restores balance via `balance.Restore(days)` |
| Cancel approved leave (`end_date < today`) | **No** — returns 400 |
| Reject pending leave (any date) | Yes |

### Balance Restoration
When approved leave is cancelled, the consumed balance is restored via `balance.Restore()`. Unlimited leave types skip balance operations.

## Summary

```
                    ┌─────────────────────┐
                    │   Admin Correction  │ ← always wins
                    └──────────┬──────────┘
                               │ overrides
                    ┌──────────▼──────────┐
                    │  Daily Attendance   │ ← source='correction'
                    └──────────┬──────────┘
                               │ guarded by
              ┌────────────────┴────────────────┐
              │                                  │
    ┌─────────▼──────────┐            ┌──────────▼────────┐
    │  ProcessDaily      │            │  ProcessRange     │
    │  (reads)           │            │  (scheduler)      │
    │  returns early if  │            │  skips via        │
    │  source=correction │            │  NOT EXISTS       │
    └────────────────────┘            └───────────────────┘
```
