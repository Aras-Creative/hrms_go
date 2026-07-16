CREATE TABLE IF NOT EXISTS job_runs (
    id        BIGSERIAL PRIMARY KEY,
    target    VARCHAR(50) NOT NULL,
    run_date  DATE NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(target, run_date)
);
