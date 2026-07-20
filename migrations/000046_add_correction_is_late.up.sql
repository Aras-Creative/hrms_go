ALTER TABLE attendance_corrections ADD COLUMN IF NOT EXISTS is_late BOOLEAN;
ALTER TABLE attendance_corrections ADD COLUMN IF NOT EXISTS is_early_leave BOOLEAN;
