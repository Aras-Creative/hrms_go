# Taste (Continuously Learned by [CommandCode][cmd])

[cmd]: https://commandcode.ai/

# workflow
- Implement contract lifecycle as draft-first: create as draft, then process (sign/send) in a separate step when admin finalizes. Confidence: 0.65
- FE should not be required to send content hashes for signing; signing validation is server-side only. Confidence: 0.70
- In seed data builder scripts, fetch sequentially over a range of IDs (1-120) and skip errors instead of aborting on failure. Confidence: 0.65

# api-design
- Group API routes by component sub-paths (e.g., `/payroll/salary/*`, `/payroll/compensation/*`) instead of flat paths like `/payroll/base-salaries`. Confidence: 0.70
- Use short names like `assignments` instead of `employee-assignments` for route path segments. Confidence: 0.65
- Keep overview endpoints summary-only; serve breakdown/detail components in separate employee-filtered endpoints. Confidence: 0.65
- Exclude heavy HTML block content (templates/blocks) from template response payloads to reduce load when pre-filling forms. Confidence: 0.65
- Do not add API override fields for data provided by other internal modules; let the UI compose the final data before calling the API instead. Confidence: 0.82

# go
- Map entities to dedicated response structs with json tags before returning in HTTP handlers, instead of serializing entities directly. Confidence: 0.60
- In JSON responses, expose profile photo URL instead of raw photo ID; keep the ID in entities/internal layers only. Confidence: 0.70
- Source salary from the employee entity, not from a salary field on the contract entity; avoid contract-level salary overrides/fallbacks. Confidence: 0.70

# validation
- Use non-blocking warnings instead of hard rejections for validation conflicts (e.g., active contracts on bulk create); let the admin decide to proceed or cancel. Confidence: 0.75

# api-design
- Accept employee IDs as an array for bulk contract creation instead of single-employee contract endpoints. Confidence: 0.65
- Consolidate single-day and date-range operations into one endpoint using identical from/to dates, instead of maintaining separate single and range endpoints. Confidence: 0.65
- Expose validation checks (e.g., active contract detection) as independent endpoints with `employee_ids[]` body param, separate from the main create flow. Confidence: 0.65

# workflow
- Generate and store contract PDF after all parties have signed, not during individual signing steps. Confidence: 0.65
- Implement contract status lifecycle as draft → sent → active: first party signs transitions draft→sent (ready for second party), second party signs transitions sent→active. Confidence: 0.70
- Keep the audit system centralized (internal/audit/) as a cross-cutting concern with sync Log() calls rather than per-module audit tables or async event bus. Confidence: 0.75

# notification
- Use broad slash-separated notification type categories (e.g., "attendance/time", "leave", "payroll", "security", "general") instead of detailed dot-notation event types. Confidence: 0.70

# api-design
- Provide lightweight pending-count endpoints (e.g., GET /contracts/me/pending → {pending: N}) for banner/notification use cases, separate from full list endpoints. Confidence: 0.70
- Expose employee-scoped self-service endpoints as /{resource}/me and /{resource}/me/{sub} that resolve JWT user_id to employee_id internally. Confidence: 0.65

# audit
- Snapshot actor name at creation time from user.full_name (user table), not resolved on the fly or from employee table. Confidence: 0.75
- Expose audit history alongside detail responses as {resource, history[]} rather than requiring separate audit endpoint calls. Confidence: 0.70
- Add ListByResource method to audit usecase for fetching all audit entries for a specific resource. Confidence: 0.70
- For enriched audit query results (e.g., actor name), use a separate read model with DB-level JOINs instead of adding enrichment fields to the domain entity. Confidence: 0.70
- Accept DTO/plain data values instead of domain entities in the audit Log() method, keeping the audit module decoupled from other modules' domain types. Confidence: 0.90

# notification
- Use a single unified SSE endpoint (GET /events) that subscribes to all relevant topics (punches, notifications:{userID}) instead of per-module SSE streams. Confidence: 0.70
- Notify all admin users (role = admin/super) on key employee actions (e.g., leave submission) via role-based FindAdminIDs query, excluding the actor themselves. Confidence: 0.70
- Use bahasa Indonesia for notification message bodies (e.g., "mengajukan cuti", "menyetujui cuti"). Confidence: 0.65
- Resolve user display names via a dedicated UserNameResolver adapter wrapping the auth UserRepository, keeping name resolution separate from employee data. Confidence: 0.65

# audit
- Forward all AuditLogData fields (including IP and UserAgent) through adapter implementations; do not drop fields when translating between audit module and other modules. Confidence: 0.75

# documentation
- Add explicit explanatory comments at business-rule decision points (e.g., "correction always wins" early returns) so future developers don't mistake deliberate logic for bugs. Confidence: 0.90

# domain-design
- Use domain entity methods instead of direct field assignment in usecase layer to avoid anemic domain model. Confidence: 0.65
- Use domain service/value objects (e.g., BankAccount value object) instead of raw primitive fields on entities for conceptually grouped data. Confidence: 0.60
- Do not import domain entities from other modules (e.g., auth entity in attendance); use adapter interfaces instead to avoid leaking domain boundaries across modules. Confidence: 0.87
- Cache the singleton settings entity in memory (e.g., via sync.Map or atomic.Value) instead of reading from DB on every request. Confidence: 0.70

# code-organization
- Extract shared utility functions (e.g., parseDateRange, time helpers) into a shared `pkg` package rather than placing them in a module's internal `models.go` file. Confidence: 0.60

# api-design
- Make API request fields optional/omitempty at the transport layer; push business-rule validation to the entity layer instead of forcing all fields required at the API boundary. Confidence: 0.65
- Split module code by aggregate (submission.go, type.go, balance.go) across service and repository layers (usecase, postgres) instead of keeping monolithic files per layer. Confidence: 0.70
- Keep delivery/handler layer in a single handler file; do not split handlers by aggregate. Confidence: 0.75
- Name Go files with just the aggregate name (e.g., submission.go, type.go, balance.go) since the module/package name already provides context; avoid redundant prefix/suffix. Confidence: 0.70
- Name adapter/interface files as port.go instead of fetcher.go to follow ports-and-adapters naming convention. Confidence: 0.65

# code-organization
- Place response/request mapping functions in dedicated response.go/request.go files inside the delivery package, not in the handler.go file. Confidence: 0.70

# code-organization
- Delivery layer should only depend on model or entity types, never on repository types. Confidence: 0.70

# attendance
See [attendance/taste.md](attendance/taste.md)
