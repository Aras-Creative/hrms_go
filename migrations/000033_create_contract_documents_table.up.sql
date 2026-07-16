CREATE TABLE contract_documents (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id   UUID NOT NULL,
    document_id   UUID NOT NULL,
    content_hash  VARCHAR(64) NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_contract_documents_contract FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE,
    CONSTRAINT fk_contract_documents_document FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_contract_documents_contract_id ON contract_documents(contract_id);