CREATE TABLE IF NOT EXISTS contract_templates (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           VARCHAR(255) NOT NULL,
    contract_type  VARCHAR(10) NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    is_active      BOOLEAN NOT NULL DEFAULT true,
    data           JSONB NOT NULL DEFAULT '{}',
    templates      JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contracts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id         UUID NOT NULL,
    employee_id         UUID NOT NULL,
    number              VARCHAR(100) NOT NULL,
    start_date          TIMESTAMPTZ,
    end_date            TIMESTAMPTZ,
    contract_date_legal TIMESTAMPTZ,
    salary              VARCHAR(100) NOT NULL DEFAULT '',
    designation_title   VARCHAR(255) NOT NULL DEFAULT '',
    working_hours       VARCHAR(100) NOT NULL DEFAULT '',
    status              VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contracts_template_id ON contracts(template_id);
CREATE INDEX IF NOT EXISTS idx_contracts_employee_id ON contracts(employee_id);
