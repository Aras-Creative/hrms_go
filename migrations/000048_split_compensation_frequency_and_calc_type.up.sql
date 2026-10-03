-- Split the overloaded employee_compensations.frequency column in two.
--
-- 'frequency' used to carry two unrelated questions at once:
--
--   * when is this paid?          -> monthly | yearly | one_time
--   * how is the amount derived?  -> per_attended_day
--
-- The deduction side already separates these: deduction_types.deduction_type holds the
-- formula and deduction_types.value_source holds where the rate comes from. This gives
-- compensations the same shape: frequency stays cadence, calc_type carries the formula.
--
-- No payslip amount changes. A per_attended_day allowance was already evaluated as
-- amount x attended_days for the period; it is now described by (frequency='monthly',
-- calc_type='per_attended_day') instead of the misleading frequency alone. 'monthly' is
-- the faithful cadence for a daily allowance paid out on a payroll run.

ALTER TABLE employee_compensations
    ADD COLUMN IF NOT EXISTS calc_type VARCHAR(20) NOT NULL DEFAULT 'fixed';

-- Carry the existing rows over before the CHECK on frequency is narrowed, so no
-- intermediate state has a row rejected by a constraint.
UPDATE employee_compensations
SET calc_type = 'per_attended_day'
WHERE frequency = 'per_attended_day';

UPDATE employee_compensations
SET frequency = 'monthly'
WHERE frequency = 'per_attended_day';

-- Recreate the CHECK so the two axes are enforced separately. The old constraint is
-- dropped by name because 000047 added it under exactly this name.
ALTER TABLE employee_compensations
    DROP CONSTRAINT IF EXISTS employee_compensations_frequency_check,
    DROP CONSTRAINT IF EXISTS employee_compensations_calc_type_check;

ALTER TABLE employee_compensations
    ADD CONSTRAINT employee_compensations_frequency_check
        CHECK (frequency IN ('monthly', 'yearly', 'one_time')),
    ADD CONSTRAINT employee_compensations_calc_type_check
        CHECK (calc_type IN ('fixed', 'per_attended_day'));

COMMENT ON COLUMN employee_compensations.frequency IS
    'How often the allowance is paid: monthly, yearly or one_time. Says nothing about how the amount is derived';
COMMENT ON COLUMN employee_compensations.calc_type IS
    'How the allowance is derived: fixed (amount once per period) or per_attended_day (amount as a daily rate times days actually present)';
COMMENT ON COLUMN employee_compensations.amount IS
    'Flat amount in cents per period, or the daily rate in cents when calc_type = per_attended_day';