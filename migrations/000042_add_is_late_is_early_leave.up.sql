ALTER TABLE daily_attendances
  ADD COLUMN is_late BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN is_early_leave BOOLEAN NOT NULL DEFAULT false;

-- Migrate existing rows: status='late' → status='present', is_late=true
UPDATE daily_attendances SET status = 'present', is_late = true WHERE status = 'late';

-- Migrate existing rows: status='early_leave' → status='present', is_early_leave=true
UPDATE daily_attendances SET status = 'present', is_early_leave = true WHERE status = 'early_leave';
