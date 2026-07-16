CREATE TABLE IF NOT EXISTS leave_submissions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id   UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    start_date    DATE NOT NULL,
    end_date      DATE NOT NULL,
    days          INT NOT NULL,
    reason        TEXT NOT NULL DEFAULT '',
    attachment_id UUID,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    approved_by   UUID REFERENCES users(id),
    approved_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leave_submissions_employee ON leave_submissions (employee_id);
CREATE INDEX idx_leave_submissions_status ON leave_submissions (status);
