-- A per-slip display label for the base salary line. The PDF previously hardcoded
-- "Gaji Pokok", so a scheme that pays a different basic (for example CRM) could not say so
-- on the slip without a code change.
--
-- Empty means "use the default": the renderer falls back to "Gaji Pokok", which keeps every
-- existing slip and every processor insert unchanged. The label is descriptive only and is
-- never added to any total.
ALTER TABLE pay_slips
    ADD COLUMN base_salary_label VARCHAR(255) NOT NULL DEFAULT '';
