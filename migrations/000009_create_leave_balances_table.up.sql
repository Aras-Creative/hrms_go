CREATE TABLE IF NOT EXISTS leave_balances (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id UUID NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    year          INT NOT NULL,
    total_days    INT NOT NULL,
    used_days     INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, leave_type_id, year)
);

CREATE INDEX idx_leave_balances_employee ON leave_balances (employee_id);
CREATE INDEX idx_leave_balances_type ON leave_balances (leave_type_id);
