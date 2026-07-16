CREATE TABLE settings (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timezone           VARCHAR(64)  NOT NULL DEFAULT 'Asia/Jakarta',
    company_name       VARCHAR(255) NOT NULL DEFAULT '',
    company_address    TEXT         NOT NULL DEFAULT '',
    company_logo_id    UUID REFERENCES documents(id) ON DELETE SET NULL,
    whitelist_ip_cidrs JSONB        NOT NULL DEFAULT '[]'::jsonb,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed the singleton row so the app always has a setting to read
INSERT INTO settings (timezone, company_name, company_address)
VALUES ('Asia/Jakarta', '', '');
