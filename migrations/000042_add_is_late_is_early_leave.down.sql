-- Restore is_late=true rows back to status='late'
UPDATE daily_attendances SET status = 'late' WHERE is_late = true;
-- Restore is_early_leave=true rows back to status='early_leave'
UPDATE daily_attendances SET status = 'early_leave' WHERE is_early_leave = true;

ALTER TABLE daily_attendances
  DROP COLUMN is_late,
  DROP COLUMN is_early_leave;
