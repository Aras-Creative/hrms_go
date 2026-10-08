ALTER TABLE leave_types
    ADD COLUMN IF NOT EXISTS include_in_attendance_allowance BOOLEAN NOT NULL DEFAULT false;

-- Set default true for standard annual leave and sick leave
UPDATE leave_types
SET include_in_attendance_allowance = true
WHERE LOWER(name) IN ('cuti', 'cuti tahunan', 'sakit', 'sick', 'sick leave');

COMMENT ON COLUMN leave_types.include_in_attendance_allowance IS
    'Whether days taken under this leave type count toward attendance-based allowances (per_attended_day)';
