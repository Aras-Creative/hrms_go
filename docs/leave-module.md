# Leave Module

Leave management — types, balances, submissions, approvals, and attendance reprocessing.

---

## Module Structure

```
internal/leave/
├── adapter/
│   ├── attendance.go            # AttendanceProcessorAdapter → reprocess attendance after leave changes
│   ├── employee_fetcher.go      # EmployeeFetcherAdapter → employee lookup, user mapping
│   ├── storage_downloader.go    # StorageAttachmentResolver → attachment URL resolution
│   └── user_name.go             # UserNameAdapter → user names, admin IDs
├── delivery/
│   ├── handler.go               # HTTP handlers (submit, approve, reject, cancel, etc.)
│   ├── request.go               # Request DTOs with validation tags
│   ├── response.go              # Response DTOs + converter functions
│   └── routes.go                # 13 routes under /leave
├── entity/
│   ├── balance.go               # LeaveBalance — quota enforcement, consume/restore
│   ├── status.go                # LeaveStatus enum (pending/approved/rejected/cancelled)
│   ├── submission.go            # LeaveSubmission — status transitions (approve/reject/cancel)
│   └── type.go                  # LeaveType — configuration (paid, unlimited, default days)
├── models/
│   ├── balance.go               # LeaveBalance DTO + input/result structs
│   ├── date.go                  # ParseDate helper (YYYY-MM-DD or RFC3339)
│   ├── submission.go            # LeaveSubmission DTO + input/result structs
│   └── type.go                  # LeaveType DTO + input structs
├── repository/
│   ├── balance.go               # PostgresLeaveBalanceRepo — balance CRUD + atomic consume
│   ├── models.go                # DB models (LeaveTypeModel, LeaveBalanceModel, etc.)
│   ├── repository.go            # 3 interfaces: LeaveType, LeaveBalance, LeaveSubmission
│   ├── submission.go            # PostgresLeaveSubmissionRepo — submission CRUD + overlap check
│   └── type.go                  # PostgresLeaveTypeRepo — type CRUD
├── usecase/
│   ├── balance.go               # ListBalances, UpdateBalance, GetEmployeeBalance
│   ├── leave.go                 # LeaveUsecase struct + constructor + port interfaces
│   ├── submission.go            # Submit, Approve, Reject, Cancel, List, Get
│   └── type.go                  # Create, Update, Delete, SeedBalances, AssignEmployeeBalances
```

---

## Routes

All under `/api/v1/`. Auth-protected unless noted.

### Leave Types

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/leave/types` | — | List all active leave types |
| GET | `/leave/types/:id` | — | Get leave type detail |
| GET | `/leave/type-options` | — | List leave types as select options |
| POST | `/leave/types` | auth | Create leave type |
| PUT | `/leave/types/:id` | auth | Update leave type |
| DELETE | `/leave/types/:id` | auth | Soft-delete leave type (disable) |

### Leave Balances

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/leave/balances` | auth | Get employee balance for a leave type |
| GET | `/leave/admin/balances` | auth+admin | List all balances (paginated, filterable) |
| PUT | `/leave/admin/balances` | auth+admin | Update employee leave balance |

### Leave Submissions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/leave/submissions` | auth | Submit leave request |
| GET | `/leave/submissions` | auth | List my submissions (paginated) |
| GET | `/leave/submissions/:id` | auth | Get submission detail + audit history |
| GET | `/leave/submissions/:id/attachment` | auth | Get attachment URL |
| POST | `/leave/submissions/:id/cancel` | auth | Cancel my submission |
| GET | `/leave/admin/submissions` | auth+admin | List all submissions (paginated, filterable) |
| POST | `/leave/admin/submissions/:id/approve` | auth+admin | Approve submission |
| POST | `/leave/admin/submissions/:id/reject` | auth+admin | Reject submission |

---

## Entity Methods

All state transitions are encapsulated in entity methods. No direct field mutation in usecases.

### LeaveSubmission

| Method | Status Guard | Transition |
|--------|-------------|------------|
| `Approve(approvedBy)` | `pending` | → `approved`, sets `ApprovedBy` + `ApprovedAt` |
| `Reject()` | `pending` | → `rejected` |
| `Cancel()` | `pending` or `approved` | → `cancelled` |

### LeaveBalance

| Method | Guard | Effect |
|--------|-------|--------|
| `Consume(days)` | `days > 0`, `UsedDays + days <= TotalDays` | Increments `UsedDays` |
| `Restore(days)` | `days > 0` | Decrements `UsedDays` (floor 0) |
| `SufficientQuota(days)` | `days > 0`, `UsedDays + days <= TotalDays` | Returns `bool` |
| `Remaining()` | none | Returns `TotalDays - UsedDays` |

### LeaveType

| Method | Effect |
|--------|--------|
| `Rename(name)` | Updates `Name` |
| `Disable()` | Sets `IsActive = false` |
| `Enable()` | Sets `IsActive = true` |

---

## State Machine

```
                    ┌──────────┐
                    │ pending  │
                    └────┬─────┘
             ┌───────────┼───────────┐
             ▼           ▼           ▼
         ┌────────┐ ┌──────────┐ ┌────────────┐
         │approved│ │ rejected │ │ cancelled  │
         └───┬────┘ └──────────┘ └────────────┘
             │
             ▼
         ┌────────────┐
         │ cancelled  │
         └────────────┘
```

- `pending` → `approved`, `rejected`, or `cancelled`
- `approved` → `cancelled` only (employee can cancel approved leave before it passes)
- `rejected` → no transitions
- `cancelled` → no transitions

---

## Submit Flow

```
Employee submits leave request
─────────────────────────────
  SubmitLeave(input)
  ├─ employeeFetcher.FindByUserID(userID) → employeeID
  ├─ leaveTypeRepo.FindByID(typeID) → validate active
  ├─ Validate: start_date not in past, end_date >= start_date
  ├─ submissionRepo.HasOverlap(employeeID, start, end) → validate no overlap
  ├─ countWeekdays(start, end) → days
  ├─ If !IsUnlimited: balanceRepo.FindByEmployeeAndTypeYear → SufficientQuota(days)
  ├─ NewLeaveSubmission(...) → status = "pending"
  └─ submissionRepo.Create(...)
```

---

## Approve Flow

```
Admin approves submission
─────────────────────────
  ApproveSubmission(submissionID, approvedBy)
  ├─ submissionRepo.FindByID(id) → submission
  ├─ submission.Approve(approvedBy) → status = "approved"
  ├─ If IsUnlimited:
  │  └─ submissionRepo.Update(...)
  ├─ If !IsUnlimited:
  │  ├─ leaveBalanceRepo.FindByEmployeeAndTypeYear → balance
  │  └─ leaveBalanceRepo.ConsumeBalance(balance.ID, days) → atomic UPDATE with WHERE
  ├─ submissionRepo.Update(...)
  └─ reprocessAttendance(employeeID, startDate, endDate)
     └─ For each date: attendanceProcessor.ReprocessDay(...)
```

---

## Cancel Flow

```
Employee cancels own submission
───────────────────────────────
  CancelSubmission(submissionID, userID)
  ├─ submissionRepo.FindByID(id) → submission
  ├─ employeeFetcher.FindByUserID(userID) → employeeID
  ├─ Validate: submission.EmployeeID == employeeID (ownership check)
  ├─ Validate: if approved + end_date < today → cannot cancel past leave
  ├─ wasApproved = (status == "approved")
  ├─ submission.Cancel() → status = "cancelled"
  ├─ If wasApproved && !IsUnlimited:
  │  ├─ leaveBalanceRepo.FindByEmployeeAndTypeYear → balance
  │  └─ balance.Restore(days) → balanceRepo.Update(...)
  ├─ submissionRepo.Update(...)
  └─ reprocessAttendance(employeeID, startDate, endDate)
```

---

## Balance Seeding

When a new leave type is created with `IsUnlimited = false` and `DefaultDays > 0`:

1. `SeedBalancesForLeaveTypeAsync` fires in background goroutine
2. `GetAllActiveIDs()` → all active employees
3. For each: `NewLeaveBalance(empID, typeID, year, defaultDays)`
4. `balanceRepo.Create(...)` for each

When a new employee is onboarded:

1. `AssignEmployeeBalances(employeeID)` called
2. For each active leave type (not unlimited, defaultDays > 0):
3. Check if balance already exists for current year
4. If not: `NewLeaveBalance(...)` → `balanceRepo.Create(...)`

---

## Cross-Module Dependencies

Leave interacts with 3 other modules via port interfaces:

| Interface | Methods | Purpose |
|-----------|---------|---------|
| `EmployeeFetcher` | `GetAllActiveIDs`, `ExistsByID`, `FindByUserID`, `FindUserIDByEmployeeID` | Employee lookup, user↔employee mapping |
| `UserNameResolver` | `FindNameByID`, `FindAdminIDs` | Actor names for audit, admin notification targets |
| `AttendanceReprocessor` | `ReprocessDay` | Re-evaluate attendance after leave status changes |

### Attendance Reprocessing

After any leave status change (submit, approve, reject, cancel), the affected dates are reprocessed:

1. For each date in `[startDate, endDate]`: `attendanceProcessor.ReprocessDay(employeeID, date)`
2. If attendance record has `source = "correction"` → skipped (manual correction wins)
3. Otherwise: attendance is recalculated based on new leave status

---

## Adapter Implementations

| Adapter | Implements | Backed By |
|---------|-----------|-----------|
| `EmployeeFetcherAdapter` | `usecase.EmployeeFetcher` | `EmployeeRepository` |
| `UserNameAdapter` | `usecase.UserNameResolver` | `auth.UserRepository` |
| `AttendanceProcessorAdapter` | `usecase.AttendanceReprocessor` | `attendance.DailyProcessor` + `EmployeeRepository` |
| `StorageAttachmentResolver` | `delivery.AttachmentResolver` | `storage.DocumentRepository` + `storage.URLResolver` |

---

## Migration History

| Migration | Description |
|-----------|-------------|
| `000008` | Create `leave_types`, `leave_balances`, `leave_submissions` tables |

---

## Architecture Notes

- **Entity methods enforce state machine**: All status transitions go through guarded methods (`Approve`, `Reject`, `Cancel`). Direct field mutation is prohibited.
- **Atomic balance consume**: `ConsumeBalance` uses `UPDATE ... SET used_days = used_days + $1 WHERE id = $2 AND used_days + $1 <= total_days` to prevent race conditions.
- **Balance restore on cancel**: When an approved submission is cancelled, the consumed days are restored via `balance.Restore(days)`.
- **Weekday counting**: `countWeekdays` skips Saturdays and Sundays — only business days consume quota.
- **Attendance reprocessing**: Leave changes trigger re-evaluation of attendance records for affected dates, with manual corrections taking priority.
- **Async balance seeding**: New leave type creation seeds balances for all active employees in a background goroutine.
- **Audit + notification**: Submit, approve, reject, and cancel actions are logged via `AuditUsecase` and trigger notifications via `NotificationUsecase`.
