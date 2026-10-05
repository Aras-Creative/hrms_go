-- Restores the validity window. The column is NOT NULL again, so every row is given a
-- sensible window: effective from the day the rollback runs, open-ended.
--
-- The original per-row dates are NOT recovered. They were dropped by the up migration and
-- nothing here can reconstruct them, so an assignment that used to apply to only October
-- now applies from the rollback date onwards, which is a wider window than it had.
--
-- That matters because the window decided which periods an item paid out on. If the exact
-- dates are needed, restore the database from a backup taken before the up migration
-- rather than relying on this.
ALTER TABLE employee_compensations ADD COLUMN IF NOT EXISTS effective_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE employee_compensations ADD COLUMN IF NOT EXISTS end_date DATE;
ALTER TABLE employee_benefits      ADD COLUMN IF NOT EXISTS effective_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE employee_benefits      ADD COLUMN IF NOT EXISTS end_date DATE;
ALTER TABLE employee_deductions    ADD COLUMN IF NOT EXISTS effective_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE employee_deductions    ADD COLUMN IF NOT EXISTS end_date DATE;