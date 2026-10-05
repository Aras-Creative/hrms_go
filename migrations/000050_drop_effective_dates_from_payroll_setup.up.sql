-- Removes the validity window from payroll setup items.
--
-- effective_date and end_date were not informational: they decided which payroll periods
-- an item applied to. A deduction with effective_date in November never appeared on an
-- October slip, and one with an end_date stopped appearing afterwards. The query was
--
--   effective_date <= period_end AND (end_date IS NULL OR end_date >= period_start)
--
-- Every assignment now applies to every period instead, which is what setup is for: an
-- employee's allowance or deduction starts when it is entered and lasts as long as it is
-- there. To change an amount, edit it; to stop one, remove it. There is no longer a way
-- to say "only October" or "from November", which is the deliberate cost of this.
--
-- employee_base_salaries is untouched on purpose. A salary is the one item whose history
-- matters: it carries forward through the per-day and percentage calculations, so keeping
-- its window is what stops a raise in November from silently repricing October when a
-- period is recalculated. Scope here was allowances, benefits and deductions.
ALTER TABLE employee_compensations DROP COLUMN IF EXISTS effective_date;
ALTER TABLE employee_compensations DROP COLUMN IF EXISTS end_date;
ALTER TABLE employee_benefits      DROP COLUMN IF EXISTS effective_date;
ALTER TABLE employee_benefits      DROP COLUMN IF EXISTS end_date;
ALTER TABLE employee_deductions    DROP COLUMN IF EXISTS effective_date;
ALTER TABLE employee_deductions    DROP COLUMN IF EXISTS end_date;