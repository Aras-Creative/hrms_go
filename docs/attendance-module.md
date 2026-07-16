# Attendance Module

Attendance tracking — punch in/out, daily computation, corrections, and scheduled processing.

---

## Module Structure

```
internal/attendance/
├── adapter/
│   ├── employee.go              # EmployeeFetcherAdapter → user-to-employee resolution
│   └── leave.go                 # LeaveFetcherAdapter → approved leave check
├── delivery/
│   ├── handler.go               # HTTP handlers (punch, daily, correction, me, SSE)
│   ├── request.go               # Request DTOs (empty — inline in handler)
│   └── routes.go                # 11 routes under /attendance
├── entity/
│   ├── correction.go            # AttendanceCorrection — manual HR overrides
│   ├── daily.go                 # DailyAttendance — computed status + 7 Mark* methods
│   └── punch.go                 # Punch — raw clock-in/out records
├── repository/
│   ├── admin.go                 # FindAllPaginated — admin attendance list with employee joins
│   ├── correction.go            # PostgresCorrectionRepo — correction CRUD + WithTx
│   ├── daily.go                 # PostgresDailyAttendanceRepo — upsert + query + WithTx
│   ├── models.go                # DB models (DailyAttendanceModel, CorrectionModel, etc.)
│   ├── postgres.go              # PostgresPunchRepo — punch CRUD
│   ├── processor.go             # ComputeDaily/ComputeRange — raw SQL computation queries
│   └── repository.go            # 3 interfaces: DailyAttendance, Punch, Correction
├── usecase/
│   ├── correction.go            # CorrectionUsecase — create/list/delete corrections
│   ├── daily.go                 # DailyAttendanceUsecase — query + admin list
│   ├── daily_processor.go       # DailyProcessor — ProcessDaily/ProcessRange + buildDailyAttendance
│   ├── me.go                    # MeUsecase — today's attendance for authenticated employee
│   ├── models.go                # PunchInput, PunchEvent, PunchHistoryInput DTOs
│   ├── port.go                  # 2 port interfaces: EmployeeFetcher, LeaveFetcher
│   ├── punch.go                 # PunchUsecase — punch in/out + SSE broadcast
│   └── scheduler.go             # Scheduler — periodic processing via pg_advisory_lock
```

---

## Routes

All under `/api/v1/`. Auth-protected unless noted.

### Punch

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/attendance/punch/in` | auth | Clock in |
| POST | `/attendance/punch/out` | auth | Clock out |
| GET | `/attendance/punch/today` | auth | Today's punches for employee |
| GET | `/attendance/punch/history` | auth | Punch history (date range) |

### Daily Attendance

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/attendance/daily` | auth | Query daily attendance (employee, date range) |
| GET | `/attendance/list` | auth+admin | Admin list (paginated, filterable by name/status/designation) |
| GET | `/attendance/mine` | auth | Today's computed attendance for authenticated employee |

### Corrections

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/attendance/corrections` | auth+admin | Create correction |
| GET | `/attendance/corrections` | auth+admin | List corrections (paginated) |
| DELETE | `/attendance/corrections/:id` | auth+admin | Delete correction (restores auto-computed values) |

### SSE

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/attendance/events` | auth+admin | Real-time punch events stream |

---

## Entity Methods

### DailyAttendance

| Method | Effect |
|--------|--------|
| `SetStatus(s)` | Validates + sets status |
| `MarkOnLeave(submissionID, typeName)` | → `on_leave` + sets leave fields |
| `MarkLate()` | → `late` |
| `MarkEarlyLeave()` | → `early_leave` |
| `MarkPresent()` | → `present` |
| `MarkPending()` | → `pending` (punched in, no punch out yet) |
| `MarkAbsent()` | → `absent` |
| `MarkNonWorking()` | → `non_working` |
| `IsLate()` / `LateMinutes()` | Compares `FirstPunchIn` vs `ExpectedStartTime` |
| `IsEarlyLeave()` | Compares `LastPunchOut` vs `ExpectedEndTime` |

### AttendanceStatus Enum

```
present | late | early_leave | absent | on_leave | non_working | pending
```

### PunchType Enum

```
in | out
```

---

## Status Computation Logic

`buildDailyAttendance` in `daily_processor.go` determines status from raw data:

| Condition | Status |
|-----------|--------|
| `LeaveSubmissionID != nil` | `on_leave` |
| Override exists + `is_working_day = false` | `non_working` |
| Has punch in, no punch out | `late` if late, else `pending` |
| Has punch in + punch out | `late` / `early_leave` / `present` based on expected times |
| Source is `working_pattern` or `override` (no punches) | `absent` |
| Default | `non_working` |

---

## Correction Flow

```
Admin creates correction
─────────────────────────
  CorrectionUsecase.Create(input)
  ├─ Validate: employee_id, reason required; clock_in < clock_out; at least one field
  ├─ Begin transaction
  ├─ processor.ComputeDaily(employeeID, date) → base attendance
  ├─ NewAttendanceCorrection(...) → saves correction record
  ├─ Apply overrides: clock_in, clock_out, status
  ├─ Set source = "correction"
  ├─ dailyRepo.Upsert(da) → overwrites daily_attendances
  └─ Commit transaction
```

**Correction wins**: Once `source = "correction"`, `ProcessDaily` skips recomputation. To restore auto-computed values, delete the correction.

### Delete Correction

```
Admin deletes correction
─────────────────────────
  CorrectionUsecase.Delete(id)
  ├─ correctionRepo.FindByID(id) → validate exists
  ├─ correctionRepo.Delete(id)
  └─ processor.ProcessDaily(employeeID, date) → recomputes from scratch
```

---

## Scheduler

Background goroutine that processes daily attendance at specific minutes past midnight (WIB timezone):

| Target (minutes) | Approx Time | Purpose |
|-------------------|-------------|---------|
| 720 | 12:00 | Midday |
| 960 | 16:00 | Afternoon |
| 975 | 16:15 | Late afternoon |
| 1080 | 18:00 | End of day |
| 1380 | 23:00 | Night catch-up |

**Distributed lock**: Uses `pg_advisory_lock` + `job_runs` table with UNIQUE constraint to prevent duplicate processing across multiple instances.

```
Scheduler.tick()
├─ Check targets vs current time
├─ pg_try_advisory_lock(0x48524D53)
├─ INSERT INTO job_runs(target, run_date) ON CONFLICT DO NOTHING
├─ If rows_affected == 0 → another instance already ran
└─ processor.ProcessRange(date, date) → upserts for all employees
```

---

## SSE Events

Punch events are broadcast to the `punches` channel via `sse.Hub`:

```json
{
  "employee_id": "...",
  "punch_type": "in",
  "timestamp": "...",
  "status": "present",
  "first_punch_in": "...",
  "last_punch_out": null,
  "late_minutes": 5,
  "total_work_seconds": null
}
```

---

## Cross-Module Dependencies

| Interface | Methods | Purpose |
|-----------|---------|---------|
| `EmployeeFetcher` | `FindByUserID` | Resolve user → employee (ID + name) |
| `LeaveFetcher` | `HasApprovedLeave` | Block punch if employee is on approved leave |

### Adapter Implementations

| Adapter | Implements | Backed By |
|---------|-----------|-----------|
| `EmployeeFetcherAdapter` | `usecase.EmployeeFetcher` | `EmployeeRepository` |
| `LeaveFetcherAdapter` | `usecase.LeaveFetcher` | `leave.LeaveSubmissionRepository` |

---

## ComputeDaily SQL

The core computation query joins across 7 tables:

```
employees
  LEFT JOIN employee_work_patterns (active pattern for date)
  LEFT JOIN work_patterns (active pattern)
  LEFT JOIN work_pattern_details (day_of_week match)
  LEFT JOIN employee_schedule_overrides (date match)
  LEFT JOIN LATERAL punches (first in, last out)
  LEFT JOIN leave_submissions (approved, date in range)
  LEFT JOIN leave_types (name)
```

Returns: expected times, source, punch timestamps, total work seconds, leave info, override info.

---

## Migration History

| Migration | Description |
|-----------|-------------|
| `000002` | Create `punches` table |
| `000006` | Create `daily_attendances` table |
| `000011` | Create `attendance_corrections` table |
| `000029` | Create `job_runs` table (for scheduler dedup) |

---

## Architecture Notes

- **Correction is authoritative**: Once set, `source = "correction"` prevents auto-recomputation. Delete correction to restore.
- **Upsert for daily**: `ON CONFLICT (employee_id, date) DO UPDATE` ensures idempotent writes.
- **ProcessDaily is called after every punch**: Creates/updates the daily summary immediately.
- **Scheduler for batch processing**: Catches up missed days and updates records at key times.
- **Punch blocked on leave**: `LeaveFetcher.HasApprovedLeave` prevents punching while on approved leave.
- **SSE for real-time updates**: Admin dashboard gets live punch events without polling.
