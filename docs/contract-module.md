# Contract Module

Employment contract management — templates, creation, signing, PDF generation, and termination orchestration.

---

## Module Structure

```
internal/contract/
├── adapter/
│   ├── designation.go          # DesignationFetcherAdapter → designation names
│   ├── document.go             # ObjectStorageAdapter + DocumentMetadataAdapter → R2 storage
│   ├── employee.go             # EmployeeFetcherAdapter → employee data + profile photos
│   ├── pdf.go                  # ChromedpRenderer → HTML-to-PDF via headless Chrome
│   └── salary.go               # SalaryFetcherAdapter → employee base salaries
├── delivery/
│   ├── handler.go              # HTTP handlers (terminate, sign, preview, download, etc.)
│   ├── request.go              # Request DTOs with validation tags
│   ├── response.go             # Response DTOs + converter functions
│   └── routes.go               # 16 routes under /contracts and /contracts/templates
├── entity/
│   ├── contract.go             # Contract — status transitions + business rules
│   ├── document.go             # ContractDocument — PDF hash linkage
│   ├── render.go               # Render data structs for HTML template
│   ├── signing.go              # ContractSigning — signature records
│   ├── status.go               # ContractStatus enum (draft/sent/active/expired/terminated)
│   ├── template.go             # ContractTemplate + data/partials + Merge()
│   └── type.go                 # ContractType enum (PKWT/PKWTT)
├── models/
│   ├── contract.go             # Template input/result DTOs
│   └── contract_create.go      # Contract + signing input/result DTOs
├── repository/
│   ├── contract.go             # PostgresContractRepo — CRUD + list + number sequence
│   ├── document.go             # PostgresDocumentRepo — contract-document junction
│   ├── repository.go           # 4 interfaces: Template, Contract, Signing, Document
│   ├── signing.go              # PostgresSigningRepo — signature records
│   └── template.go             # PostgresTemplateRepo — template CRUD
├── usecase/
│   ├── contract.go             # ContractUsecase — create, list, detail, check active
│   ├── document.go             # DocumentUsecase — store PDF, download PDF
│   ├── preprocess.go           # Render helpers — field resolution, gender/religion maps
│   ├── render.go               # RenderUsecase — HTML rendering + PDF preview
│   ├── signing.go              # SigningUsecase — bulk sign + auto-expire old contract
│   ├── termination.go          # TerminationUsecase — cross-module termination orchestration
│   ├── template.go             # ContractUsecase — template CRUD + EmployeeFetcher interface
│   └── templates/              # Embedded HTML template + company logo
```

---

## Routes

All under `/api/v1/`. Auth-protected unless noted.

### Templates

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/contracts/templates` | auth | List templates (paginated, filterable) |
| GET | `/contracts/templates/:id` | — | Get template detail |
| GET | `/contracts/templates/:id/prefill` | — | Get template data for contract prefill |
| POST | `/contracts/templates` | auth | Create template |
| PUT | `/contracts/templates/:id` | auth | Update template |
| DELETE | `/contracts/templates/:id` | auth | Delete template |

### Contracts

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/contracts` | auth | List contracts (paginated, with employee briefs + signing timestamps) |
| GET | `/contracts/me` | auth | List contracts for authenticated employee |
| GET | `/contracts/me/active` | auth | Get active contract for authenticated employee |
| GET | `/contracts/me/pending` | auth | Count pending contracts for authenticated employee |
| GET | `/contracts/:id` | auth | Get contract detail (with template name, signings) |
| GET | `/contracts/:id/preview` | auth | Preview contract PDF (unsigned) |
| GET | `/contracts/:id/download` | auth | Download signed contract PDF |
| POST | `/contracts` | auth | Create contracts (bulk, per employee) |
| POST | `/contracts/active` | auth | Check which employees have active contracts |
| POST | `/contracts/:id/generate-pdf` | auth | Manually generate final signed PDF |
| POST | `/contracts/:id/terminate` | auth | Terminate contract + employee + auth + schedule |

### Signing

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/contracts/sign` | auth | Sign as first party (admin) |
| POST | `/contracts/sign/second-party` | auth | Sign as second party (employee, ownership validated) |

---

## Entity Methods

All state transitions are encapsulated in entity methods. No direct field mutation in usecases.

| Method | Status Guard | Transition |
|--------|-------------|------------|
| `MarkSent()` | `draft` | → `sent` |
| `MarkActive()` | `draft` or `sent` | → `active` |
| `CanSign(party)` | `draft` (first), `draft`/`sent` (second) | validates signing allowed |
| `EvaluateSigningState(signings)` | any | calls `MarkSent`/`MarkActive` as needed, returns `shouldGeneratePDF` |
| `Expire()` | `active` | → `expired` |
| `Terminate()` | `active` | → `terminated` |
| `AttachDocument()` | any | updates `UpdatedAt` |
| `AddSignature(...)` | any | creates `ContractSigning` record |

---

## Signing Flow

```
Admin signs (first party)          Employee signs (second party)
─────────────────────────          ──────────────────────────────
  BulkSign(party="first")            BulkSignAsSecondParty(userID)
  ├─ CanSign("first")?              ├─ Resolve userID → employeeID
  ├─ AddSignature()                 ├─ Validate contract ownership
  ├─ EvaluateSigningState()         ├─ CanSign("second")?
  │  └─ hasFirst → MarkSent()       ├─ AddSignature()
  ├─ UpdateContract()               ├─ EvaluateSigningState()
  └─ return                         │  └─ hasFirst+hasSecond → MarkActive()
                                    ├─ UpdateContract()
                                    ├─ StorePDF() → R2 + contract_documents
                                    ├─ FindActiveByEmployeeID()
                                    │  └─ oldContract.Expire()
                                    │  └─ UpdateContract(oldContract)
                                    └─ return
```

---

## Contract Documents

PDFs are stored in R2 object storage with a content hash for integrity.

```
contracts/signings (both parties signed)
    │
    ▼
DocumentUsecase.StorePDF()
    ├─ RenderUsecase.Preview() → HTML → chromedp → PDF bytes
    ├─ sha256(pdfBytes) → contentHash
    ├─ R2 upload → contracts/{year}/{month}/{docID}.pdf
    ├─ documents table (storage metadata, uploaded_by=NULL)
    └─ contract_documents junction (contract_id, document_id, content_hash)
```

Download resolves: `contract_documents` → `documents` → R2 presigned URL.

Filename format: `Perjanjian Kerja - {contractNumber}.pdf`

---

## Termination Flow

`POST /contracts/:id/terminate` with `{ "termination_date": "2026-07-12" }`

Orchestrates 5 cross-module side effects:

| Step | Module | Action |
|------|--------|--------|
| 1 | Contract | `status` → `terminated` |
| 2 | Employee | `status` → `inactive`, `is_active` → `false`, `termination_date` set |
| 3 | Auth | `users.is_active` → `false` |
| 4 | Auth | All sessions revoked (`is_active` → `false`) |
| 5 | Auth | All devices revoked (`is_active` → `false`) |
| 6 | Schedule | Work pattern deactivated (`valid_to` = termination date) |
| 7 | Schedule | Future overrides deleted (`date >= termination date`) |

All steps are transactional at the usecase level — if any critical step fails, the entire operation returns an error.

---

## Auto-Expire Old Contract

When the second party signs and both parties have signed (new contract → `active`):

1. Query: `SELECT ... FROM contracts WHERE employee_id = $1 AND status = 'active'`
2. If found and ID differs from new contract → `oldContract.Expire()` → `status` → `expired`
3. Old contract is superseded, new contract is the active one

---

## Migration History

| Migration | Description |
|-----------|-------------|
| `000022` | Create `contract_templates` + `contracts` tables |
| `000023` | Add `data` + `templates` JSONB columns |
| `000024` | Create `contract_signings` table |
| `000025` | Add `sent_at` to contracts |
| `000027` | Add `designation_id` to contracts |
| `000028` | Add `party`, `signed_by_name`, `signed_by_title`, `place` to signings |
| `000033` | Create `contract_documents` junction table |
| `000034` | Drop unused columns (`contract_date_legal`, `working_hours`) |
| `000035` | Make `documents.uploaded_by` nullable |
| `000036` | Add `termination_date` to employees |

---

## Usecase Interfaces

Cross-module dependencies are expressed as narrow interfaces in the usecase layer:

| Interface | Defined In | Methods | Implemented By |
|-----------|-----------|---------|----------------|
| `EmployeeFetcher` | `template.go` | `FindByID`, `FindDesignationIDs`, `FindEmployeeIDByUserID`, `FindUserIDByEmployeeID`, `FindBriefByIDs` | `EmployeeFetcherAdapter` |
| `DesignationFetcher` | `template.go` | `FindNamesByIDs` | `DesignationFetcherAdapter` |
| `SalaryFetcher` | `template.go` | `FindCurrentByEmployeeIDs` | `SalaryFetcherAdapter` |
| `PDFRenderer` | `pdf.go` | `Render` | `ChromedpRenderer` |
| `ObjectStorage` | `document.go` | `Upload`, `Download`, `Delete` | `ObjectStorageAdapter` |
| `DocumentMetadataRepository` | `document.go` | `CreateForModule`, `FindByID` | `DocumentMetadataAdapter` |
| `ContractFinder` | `termination.go` | `FindContractByID`, `UpdateContract` | `PostgresContractRepo` |
| `EmployeeFinderTerminate` | `termination.go` | `FindByID`, `Update` | `PostgresEmployeeRepo` |
| `UserDeactivator` | `termination.go` | `Deactivate` | `PostgresUserRepo` |
| `SessionRevoker` | `termination.go` | `RevokeAllByUserID` | `PostgresSessionRepo` |
| `DeviceRevoker` | `termination.go` | `RevokeDeviceByUserID` | `PostgresDeviceRepo` |
| `WorkPatternDeactivator` | `termination.go` | `DeactivateCurrent` | `PostgresEmployeeWorkPatternRepo` |
| `ScheduleOverrideDeleter` | `termination.go` | `DeleteFutureOverridesByEmployee` | `PostgresScheduleOverrideRepo` |

---

## Architecture Notes

- **Entity methods enforce state machine**: All status transitions go through guarded methods (`MarkSent`, `MarkActive`, `Expire`, `Terminate`). Direct field mutation is prohibited.
- **Signature images stored as data URIs**: Frontend sends full `data:image/png;base64,...` strings. Template uses `safeURL` func to prevent URL sanitizer stripping.
- **Number generation**: Contract numbers use `numbergen.Generator` with prefix `CTR` and designation-based sequences.
- **PDF rendering**: HTML template → chromedp headless Chrome → PDF. Company logo embedded as base64.
- **Contract list includes employee briefs**: `FindBriefByIDs` resolves employee names + profile photo URLs in batch.
- **Old contract auto-expires**: When a new contract becomes active, the previous active contract for the same employee is automatically expired.
- **Termination is cross-module**: A single endpoint orchestrates contract, employee, auth, session, device, work pattern, and schedule changes.
- **Audit logging**: Create, sign (first/second party), and terminate actions are logged via `AuditUsecase`.
