CREATE TABLE IF NOT EXISTS employee_work_patterns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    work_pattern_id UUID NOT NULL REFERENCES work_patterns(id) ON DELETE RESTRICT,
    valid_from      DATE NOT NULL,
    valid_to        DATE,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, valid_from)
);

CREATE INDEX idx_employee_work_patterns_employee ON employee_work_patterns (employee_id);
CREATE INDEX idx_employee_work_patterns_dates  ON employee_work_patterns (employee_id, valid_from, valid_to);
