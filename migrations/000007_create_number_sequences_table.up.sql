CREATE TABLE IF NOT EXISTS number_sequences (
    designation_code VARCHAR(10) PRIMARY KEY,
    prefix           VARCHAR(4) NOT NULL,
    last_sequence    INT NOT NULL DEFAULT 0,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
