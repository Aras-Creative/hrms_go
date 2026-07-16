CREATE TABLE IF NOT EXISTS employee_schedule_overrides (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    is_working_day  BOOLEAN NOT NULL DEFAULT true,
    start_time      VARCHAR(5),
    end_time        VARCHAR(5),
    reason          VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, date)
);

CREATE INDEX idx_employee_schedule_overrides_employee ON employee_schedule_overrides (employee_id);
CREATE INDEX idx_employee_schedule_overrides_date ON employee_schedule_overrides (date);
