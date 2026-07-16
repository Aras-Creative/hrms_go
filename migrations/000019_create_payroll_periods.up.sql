CREATE EXTENSION IF NOT EXISTS "pgcrypto";

ALTER TABLE deduction_types ADD COLUMN slug VARCHAR(50) UNIQUE;

CREATE TABLE IF NOT EXISTS payroll_periods (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) NOT NULL,
    start_date    DATE NOT NULL,
    end_date      DATE NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'closed')),
    working_days  INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS pay_slips (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_id            UUID NOT NULL REFERENCES payroll_periods(id) ON DELETE CASCADE,
    employee_id          UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    base_salary          BIGINT NOT NULL,
    total_compensations  BIGINT NOT NULL,
    total_deductions     BIGINT NOT NULL,
    absent_deduction     BIGINT NOT NULL DEFAULT 0,
    absent_days          INT NOT NULL DEFAULT 0,
    net_salary           BIGINT NOT NULL,
    currency             VARCHAR(3) NOT NULL DEFAULT 'IDR',
    compensations_breakdown JSONB NOT NULL DEFAULT '[]',
    deductions_breakdown    JSONB NOT NULL DEFAULT '[]',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(period_id, employee_id)
);

CREATE INDEX idx_pay_slips_period ON pay_slips (period_id);
CREATE INDEX idx_pay_slips_employee ON pay_slips (employee_id);
