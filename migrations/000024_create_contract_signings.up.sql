CREATE TABLE IF NOT EXISTS contract_signings (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id      UUID NOT NULL,
    contract_id      UUID NOT NULL REFERENCES contracts(id),
    signed_by        VARCHAR(50) NOT NULL,
    signature_base64 TEXT NOT NULL DEFAULT '',
    signed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contract_signings_contract_id ON contract_signings(contract_id);
CREATE INDEX IF NOT EXISTS idx_contract_signings_document_id ON contract_signings(document_id);
