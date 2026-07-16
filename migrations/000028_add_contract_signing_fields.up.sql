ALTER TABLE contract_signings ADD COLUMN party VARCHAR(10) NOT NULL DEFAULT 'first';
ALTER TABLE contract_signings ADD COLUMN signed_by_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE contract_signings ADD COLUMN signed_by_title VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE contract_signings ADD COLUMN place VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE contract_signings ADD COLUMN content_hash VARCHAR(255) NOT NULL DEFAULT '';
