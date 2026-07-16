ALTER TABLE contracts DROP COLUMN IF EXISTS contract_date_legal;
ALTER TABLE contracts DROP COLUMN IF EXISTS working_hours;
ALTER TABLE contract_signings DROP COLUMN IF EXISTS document_id;
ALTER TABLE contract_signings DROP COLUMN IF EXISTS content_hash;
DROP INDEX IF EXISTS idx_contract_signings_document_id;