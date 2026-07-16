CREATE TABLE IF NOT EXISTS leave_types (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    default_days INT NOT NULL DEFAULT 0,
    is_paid      BOOLEAN NOT NULL DEFAULT TRUE,
    is_unlimited BOOLEAN NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leave_types_active ON leave_types (is_active);
