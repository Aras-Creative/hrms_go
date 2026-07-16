CREATE TABLE IF NOT EXISTS designations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       VARCHAR(10) NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_designations_code ON designations (code);
