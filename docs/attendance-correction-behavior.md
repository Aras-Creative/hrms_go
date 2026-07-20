# Attendance Correction: Behavior & Tradeoffs

## Priority Model (Current)

Correction = **lock**, not a priority overlay.

```
Daily attendance with no correction:
  Schedule ──► Punches ──► Leave ──► Absent eval ──► Upsert

Daily attendance with correction:
  [Locked] ──► source = "correction" ──► future sweeps/punches skip this row
```

## What Happens in Each Scenario

### Punch after a correction (the open question)

- The punch **is recorded** in the `punches` table (raw event, never lost)
- The `daily_attendance` row **does not change** — correction stays
- The correction must be **deleted** before the punch affects attendance

**Current mitigation**: Corrections are created by HR/admin with a reason. The person punching after being corrected is a rare cross-over (past dates, manual retroactive punch). The punch still exists for audit/reporting.

### Correction after a punch (normal flow)

- Punch is recorded in `punches` 
- Correction is created → `ApplyTo()` overwrites `daily_attendance` with correction values
- Source becomes `"correction"`
- The original punch remains in `punches` but does not drive the status

### Correction deleted

- The `daily_attendance` row is deleted
- `ProcessDaily` reruns from scratch → recomputes from schedule + punches + leaves
- Falls back to its natural computed state

### Sweep (scheduler) after a correction

- The scheduler's `ProcessRange` skips employees with `source = "correction"`
- Those rows are never recalculated until the correction is removed

## Tradeoffs

| | Current (lock) | Alternative (priority overlay) |
|---|---|---|
| **Complexity** | Simple — one check, skip | Need correction repo in processor, extra query per row |
| **Punch after correction** | Punched recorded but invisible in daily_attendance | Punch evaluated first, correction wins on top |
| **Schedule change after correction** | Ignored | Recalculated, correction overlays |
| **DB load** | Zero extra queries | 1 extra query per employee per sweep |
| **Mental model** | "Correction = final, delete to refresh" | "Correction = priority, always wins" |

## Is This a Problem?

For current usage patterns — **no**. Corrections are used for past dates where real-time punches don't happen. The punch-after-correction scenario is a theoretical edge case, not a live issue.

The tradeoff is reasonable today. If the system grows to support retroactive clock-in (e.g., employee self-service punches for past dates), the refactor becomes worthwhile.
