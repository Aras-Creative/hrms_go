CREATE TABLE IF NOT EXISTS daily_attendances (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id          UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date                 DATE NOT NULL,
    status               VARCHAR(20) NOT NULL,
    expected_start_time  TIME,
    expected_end_time    TIME,
    source               VARCHAR(20) NOT NULL,
    first_punch_in       TIMESTAMPTZ,
    last_punch_out       TIMESTAMPTZ,
    total_work_seconds   INT,
    leave_submission_id  UUID REFERENCES leave_submissions(id),
    leave_type_name      VARCHAR(255),
    schedule_override_id UUID REFERENCES employee_schedule_overrides(id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(employee_id, date)
);

CREATE INDEX idx_daily_attendances_employee ON daily_attendances (employee_id);
CREATE INDEX idx_daily_attendances_date ON daily_attendances (date);
CREATE INDEX idx_daily_attendances_employee_date ON daily_attendances (employee_id, date);
