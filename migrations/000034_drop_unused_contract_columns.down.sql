ALTER TABLE contract_signings ADD COLUMN IF NOT EXISTS content_hash VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE contract_signings ADD COLUMN IF NOT EXISTS document_id UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS working_hours VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS contract_date_legal TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_contract_signings_document_id ON contract_signings(document_id);