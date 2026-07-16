# attendance
- Treat "pending" as a frontend-only display flag in lists, not as a real domain status in the attendance system. Confidence: 0.65
- Detect half-day leave from an is_half_day flag on the leave type entity (not on individual leave submissions); apply half-day logic in the daily processor when determining attendance status. Confidence: 0.65
- Do not hardcode timezone in SQL queries; timezone must come from config (e.g., a configurable application setting), not hardcoded like 'Asia/Jakarta'. Confidence: 0.85
- Use UTC as the application-wide timezone. Confidence: 0.75
- For half-day leave auto clock-in/clock-out, use the approval time (time.Now()) instead of a hardcoded default hour. Confidence: 0.65
- Do not trigger attendance processing on List endpoints; processing should only run via the scheduler, not on read/query calls. Confidence: 0.75
- Use "present" status (not "pending") when employee has clocked in but not yet clocked out and isn't late. Confidence: 0.70
- Use "day_off" instead of "non_working" for days when employee is not scheduled to work (weekends, holidays, day off overrides). Confidence: 0.70
- Separate lateness into two independent dimensions: attendance status (present/absent) and punctuality (is_late: true/false), instead of conflating them into a single "late" status. Confidence: 0.70
