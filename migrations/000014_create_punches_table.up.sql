CREATE TABLE IF NOT EXISTS punches (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type        VARCHAR(10) NOT NULL CHECK (type IN ('in', 'out')),
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_punches_employee ON punches (employee_id);
CREATE INDEX idx_punches_timestamp ON punches (timestamp);
CREATE INDEX idx_punches_employee_date ON punches (employee_id, timestamp);
