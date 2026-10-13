-- Employees listed here bypass the global punch IP allow-list entirely. This
-- covers the edge case where a specific employee (for example a remote or
-- field worker) must be able to punch from an IP outside whitelist_ip_cidrs.
ALTER TABLE settings
    ADD COLUMN whitelist_ip_bypass_employee_ids JSONB NOT NULL DEFAULT '[]'::jsonb;
