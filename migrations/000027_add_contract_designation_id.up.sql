ALTER TABLE contracts ADD COLUMN designation_id UUID;
CREATE INDEX IF NOT EXISTS idx_contracts_designation_id ON contracts(designation_id);
