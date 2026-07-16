CREATE TABLE IF NOT EXISTS work_patterns (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS work_pattern_details (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_pattern_id   UUID NOT NULL REFERENCES work_patterns(id) ON DELETE CASCADE,
    day_of_week       INT NOT NULL,
    start_time        VARCHAR(5),
    end_time          VARCHAR(5),
    UNIQUE (work_pattern_id, day_of_week)
);

CREATE INDEX idx_work_pattern_details_pattern ON work_pattern_details (work_pattern_id);
