-- Attendance-based payroll components.
--
-- 1. Deductions gain a 'per_day' basis plus an explicit source for the daily rate,
--    so "potongan alpa" can be either a flat rupiah-per-day figure or derived
--    from the employee's daily wage.
-- 2. Compensation (allowance) gains a 'per_attended_day' frequency so transport
--    and meal allowances can be prorated by days actually worked.

-- --- Deduction types: per_day basis + daily rate source ---

-- Drop every CHECK guarding deduction_types.deduction_type, found via pg_constraint
-- rather than by convention name. A guessed name that misses would let this DROP no-op
-- silently and leave the old CHECK in force, which would then reject 'per_day'.
DO $$
DECLARE
    col_attnum SMALLINT;
    con_name   TEXT;
BEGIN
    SELECT attnum INTO col_attnum
      FROM pg_attribute
     WHERE attrelid = 'deduction_types'::regclass
       AND attname   = 'deduction_type'
       AND NOT attisdropped;

    FOR con_name IN
        SELECT c.conname
          FROM pg_constraint c
         WHERE c.conrelid = 'deduction_types'::regclass
           AND c.contype  = 'c'
           AND (c.conkey IS NULL OR c.conkey = ARRAY[col_attnum]::SMALLINT[])
    LOOP
        EXECUTE format('ALTER TABLE deduction_types DROP CONSTRAINT %I', con_name);
    END LOOP;
END $$;

ALTER TABLE deduction_types
    ADD COLUMN IF NOT EXISTS value_source VARCHAR(20) NOT NULL DEFAULT 'fixed',
    ADD COLUMN IF NOT EXISTS unit_amount  BIGINT      NOT NULL DEFAULT 0;

ALTER TABLE deduction_types
    DROP CONSTRAINT IF EXISTS deduction_types_deduction_type_check,
    DROP CONSTRAINT IF EXISTS deduction_types_value_source_check;

ALTER TABLE deduction_types
    ADD CONSTRAINT deduction_types_deduction_type_check
        CHECK (deduction_type IN ('percentage', 'fixed', 'per_day')),
    ADD CONSTRAINT deduction_types_value_source_check
        CHECK (value_source IN ('fixed', 'daily_wage'));

-- Per-employee override of the daily rate. NULL falls back to the type's unit_amount.
ALTER TABLE employee_deductions
    ADD COLUMN IF NOT EXISTS unit_amount BIGINT;

COMMENT ON COLUMN deduction_types.deduction_type IS
    'How the amount is derived: percentage (rate % of base salary), fixed (flat amount per period), per_day (daily rate x attendance-derived day count)';
COMMENT ON COLUMN deduction_types.value_source IS
    'Where the per_day rate comes from: fixed (unit_amount) or daily_wage (base salary / working days)';
COMMENT ON COLUMN deduction_types.unit_amount IS
    'Flat amount in cents for the per_day basis when value_source = fixed';
COMMENT ON COLUMN employee_deductions.unit_amount IS
    'Per-employee daily rate in cents; overrides deduction_types.unit_amount when set';

-- Move the legacy 'absent' deduction type onto the per_day basis.
--
-- Before this migration the calculation query deliberately excluded slug = 'absent':
--
--     AND dt.is_active = true AND (dt.slug IS NULL OR dt.slug != 'absent')
--
-- so the absent-day deduction was never derived from this row. It reached the payslip
-- through a hand-supplied flat 'absent_deduction' field instead, and the `fixed` branch
-- of the calculator was already a flat amount (value * 100), never a percentage.
-- `default_value` on this row was therefore dead data that no code path ever read.
--
-- Dropping that exclusion is the point of this migration, and per_day + daily_wage
-- reproduces the intended behaviour: absent days x (base salary / working days).
--
-- The narrowing condition is deliberate. Guarding on default_value = 100 would look
-- safer but is not: leaving a row on `fixed` after the exclusion is removed turns it
-- into a flat deduction charged once per period, which is worse than converting it.
-- Because default_value is never consulted for slug = 'absent', no rate is lost.
UPDATE deduction_types
SET deduction_type = 'per_day',
    value_source   = 'daily_wage',
    unit_amount    = 0
WHERE slug = 'absent'
  AND deduction_type = 'fixed'
  AND value_source   = 'fixed';

-- --- Compensation items: per_attended_day allowance ---

-- Same discovery approach for the frequency CHECK.
DO $$
DECLARE
    col_attnum SMALLINT;
    con_name   TEXT;
BEGIN
    SELECT attnum INTO col_attnum
      FROM pg_attribute
     WHERE attrelid = 'employee_compensations'::regclass
       AND attname   = 'frequency'
       AND NOT attisdropped;

    FOR con_name IN
        SELECT c.conname
          FROM pg_constraint c
         WHERE c.conrelid = 'employee_compensations'::regclass
           AND c.contype  = 'c'
           AND (c.conkey IS NULL OR c.conkey = ARRAY[col_attnum]::SMALLINT[])
    LOOP
        EXECUTE format('ALTER TABLE employee_compensations DROP CONSTRAINT %I', con_name);
    END LOOP;
END $$;

ALTER TABLE employee_compensations
    ADD CONSTRAINT employee_compensations_frequency_check
        CHECK (frequency IN ('monthly', 'yearly', 'one_time', 'per_attended_day'));

-- amount is reinterpreted as the per-day unit rate when frequency = per_attended_day.
COMMENT ON COLUMN employee_compensations.frequency IS
    'How often the allowance is paid; per_attended_day uses amount as the daily rate multiplied by days present';
COMMENT ON COLUMN employee_compensations.amount IS
    'Flat amount in cents per period, or the daily rate in cents when frequency = per_attended_day';