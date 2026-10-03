-- Reverts attendance-based payroll components.
--
-- WARNING: pay_slips produced while per_day / per_attended_day was active are NOT
-- recalculated back. Regenerate the affected periods after running this down.
--
-- The rollback is deliberately conservative. It restores only the rows this migration
-- owns, and refuses to proceed when it finds configuration it cannot faithfully undo,
-- rather than silently rewriting an operator's payroll setup.

-- --- Guard: report rows that cannot be reverted automatically ---

-- A per_day row other than the legacy 'absent' one was configured while this migration
-- was live. Folding it back to `fixed` would change a daily-rate deduction into a flat
-- monthly one, so stop and let a human decide.
DO $$
DECLARE
    leftovers TEXT;
BEGIN
    SELECT string_agg(format('%s (slug=%s, value_source=%s)', id, coalesce(slug, '-'), value_source), E'\n  ')
      INTO leftovers
      FROM deduction_types
     WHERE deduction_type = 'per_day'
       AND (slug IS DISTINCT FROM 'absent' OR value_source <> 'daily_wage');

    IF leftovers IS NOT NULL THEN
        RAISE EXCEPTION
            'rollback of 000047 aborted: these per_day deduction types were configured after the '
            'migration and cannot be reverted automatically:%s. Convert or disable them first.',
            E'\n  ' || leftovers;
    END IF;

    SELECT string_agg(format('%s (item=%s, employee=%s)', ci.id, ci.name, ec.employee_id), E'\n  ')
      INTO leftovers
      FROM employee_compensations ec
      JOIN compensation_items ci ON ci.id = ec.compensation_item_id
     WHERE ec.frequency = 'per_attended_day';

    IF leftovers IS NOT NULL THEN
        RAISE EXCEPTION
            'rollback of 000047 aborted: these compensations use per_attended_day and cannot be '
            'reverted automatically:%s. Change their frequency first.', E'\n  ' || leftovers;
    END IF;
END $$;

-- --- Compensation items ---

-- Only reachable when no per_attended_day rows remain (the guard above aborts otherwise).
UPDATE employee_compensations
SET frequency = 'monthly'
WHERE frequency = 'per_attended_day';

ALTER TABLE employee_compensations
    DROP CONSTRAINT IF EXISTS employee_compensations_frequency_check;

ALTER TABLE employee_compensations
    ADD CONSTRAINT employee_compensations_frequency_check
        CHECK (frequency IN ('monthly', 'yearly', 'one_time'));

-- --- Deduction types ---

-- Restore the legacy 'absent' type. default_value is left untouched on purpose: the
-- up migration never wrote to it, so whatever the operator had stored is still there.
UPDATE deduction_types
SET deduction_type = 'fixed',
    value_source   = 'fixed',
    unit_amount    = 0
WHERE slug = 'absent'
  AND deduction_type = 'per_day'
  AND value_source   = 'daily_wage';

ALTER TABLE employee_deductions
    DROP COLUMN IF EXISTS unit_amount;

ALTER TABLE deduction_types
    DROP CONSTRAINT IF EXISTS deduction_types_deduction_type_check,
    DROP CONSTRAINT IF EXISTS deduction_types_value_source_check;

ALTER TABLE deduction_types
    DROP COLUMN IF EXISTS value_source,
    DROP COLUMN IF EXISTS unit_amount;

ALTER TABLE deduction_types
    ADD CONSTRAINT deduction_types_deduction_type_check
        CHECK (deduction_type IN ('percentage', 'fixed'));