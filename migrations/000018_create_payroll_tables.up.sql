CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Employee Base Salaries
CREATE TABLE IF NOT EXISTS employee_base_salaries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    amount         BIGINT NOT NULL,
    currency       VARCHAR(3) NOT NULL DEFAULT 'IDR',
    effective_date DATE NOT NULL,
    end_date       DATE,
    notes          TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_emp_base_sal_employee ON employee_base_salaries (employee_id);
CREATE INDEX idx_emp_base_sal_effective ON employee_base_salaries (employee_id, effective_date DESC);

-- Compensation Items (master data)
CREATE TABLE IF NOT EXISTS compensation_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    item_type   VARCHAR(20) NOT NULL CHECK (item_type IN ('recurring', 'one_time')),
    description TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    is_taxable  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Employee Compensations
CREATE TABLE IF NOT EXISTS employee_compensations (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id          UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    compensation_item_id UUID NOT NULL REFERENCES compensation_items(id) ON DELETE CASCADE,
    amount               BIGINT NOT NULL,
    frequency            VARCHAR(20) NOT NULL CHECK (frequency IN ('monthly', 'yearly', 'one_time')),
    effective_date       DATE NOT NULL,
    end_date             DATE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_emp_comp_employee ON employee_compensations (employee_id);
CREATE INDEX idx_emp_comp_item ON employee_compensations (compensation_item_id);

-- Benefit Types (master data)
CREATE TABLE IF NOT EXISTS benefit_types (
    id                           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                         VARCHAR(255) NOT NULL,
    description                  TEXT NOT NULL DEFAULT '',
    employer_contribution_type   VARCHAR(20) NOT NULL CHECK (employer_contribution_type IN ('percentage', 'fixed')),
    employer_contribution_value  DECIMAL(12,2) NOT NULL DEFAULT 0,
    employee_contribution_type   VARCHAR(20) NOT NULL CHECK (employee_contribution_type IN ('percentage', 'fixed')),
    employee_contribution_value  DECIMAL(12,2) NOT NULL DEFAULT 0,
    is_active                    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Employee Benefits
CREATE TABLE IF NOT EXISTS employee_benefits (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id        UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    benefit_type_id    UUID NOT NULL REFERENCES benefit_types(id) ON DELETE CASCADE,
    participant_number VARCHAR(50) NOT NULL DEFAULT '',
    effective_date     DATE NOT NULL,
    end_date           DATE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, benefit_type_id, effective_date)
);

CREATE INDEX idx_emp_benefits_employee ON employee_benefits (employee_id);

-- Deduction Types (master data)
CREATE TABLE IF NOT EXISTS deduction_types (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    deduction_type  VARCHAR(20) NOT NULL CHECK (deduction_type IN ('percentage', 'fixed')),
    default_value   DECIMAL(12,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    is_mandatory    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Employee Deductions
CREATE TABLE IF NOT EXISTS employee_deductions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id       UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    deduction_type_id UUID NOT NULL REFERENCES deduction_types(id) ON DELETE CASCADE,
    value             DECIMAL(12,2),
    effective_date    DATE NOT NULL,
    end_date          DATE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, deduction_type_id, effective_date)
);

CREATE INDEX idx_emp_deductions_employee ON employee_deductions (employee_id);
