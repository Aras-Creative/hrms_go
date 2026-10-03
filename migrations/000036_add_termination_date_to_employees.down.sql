-- Rolls back the termination_date column added by the up migration.
--
-- Destructive by design, like every down migration that drops a column: any termination
-- dates recorded since are lost. It exists so `migrate down` can reach a consistent
-- version instead of failing on a missing file part-way through a rollback.
ALTER TABLE employees
    DROP COLUMN IF EXISTS termination_date;