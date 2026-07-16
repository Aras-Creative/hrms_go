CREATE TABLE IF NOT EXISTS employees (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID,
    full_name               VARCHAR(255) NOT NULL,
    employee_number         VARCHAR(15) NOT NULL UNIQUE,
    phone                   VARCHAR(15) NOT NULL,
    personal_email          VARCHAR(255),
    emergency_contact_name  VARCHAR(255) NOT NULL,
    emergency_contact_phone VARCHAR(15) NOT NULL,
    place_of_birth          VARCHAR(255) NOT NULL,
    date_of_birth           DATE,
    join_date               DATE,
    gender                  VARCHAR(10) NOT NULL,
    education               VARCHAR(255) NOT NULL,
    status                  VARCHAR(20) NOT NULL,
    address                 TEXT NOT NULL,
    designation_id          UUID,
    national_id             VARCHAR(50) NOT NULL,
    religion                VARCHAR(20) NOT NULL,
    profile_photo_id        UUID,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_employees_employee_number ON employees (employee_number);
CREATE INDEX idx_employees_user_id ON employees (user_id);
CREATE INDEX idx_employees_designation_id ON employees (designation_id);
