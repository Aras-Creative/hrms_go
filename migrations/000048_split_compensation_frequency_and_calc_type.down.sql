-- Reverts the split of employee_compensations.frequency into cadence plus calc_type.
--
-- WARNING: allowances created while calc_type existed lose their distinction. Rows whose
-- calc_type is not 'fixed' cannot be represented in the old single column without
-- guessing a cadence, so the rollback aborts while any such row exists rather than
-- silently flattening them.

DO $$
DECLARE
    non_default TEXT;
BEGIN
    SELECT string_agg(format('%s (calc_type=%s)', ec.id, ec.calc_type), E'\n  ')
      INTO non_default
      FROM employee_compensations ec
     WHERE ec.calc_type <> 'fixed';

    IF non_default IS NOT NULL THEN
        RAISE EXCEPTION
            'rollback of 000048 aborted: these compensations use calc_type = per_attended_day, '
            'which the old frequency column cannot express without losing the derivation.%s',
            E'\n  ' || non_default;
    END IF;
END $$;

-- Re-join the two axes, then drop the calc_type CHECK before the column goes away.
ALTER TABLE employee_compensations
    DROP CONSTRAINT IF EXISTS employee_compensations_calc_type_check;

ALTER TABLE employee_compensations
    DROP COLUMN IF EXISTS calc_type;

-- Restore the wider frequency CHECK from 000047.
ALTER TABLE employee_compensations
    DROP CONSTRAINT IF EXISTS employee_compensations_frequency_check;

ALTER TABLE employee_compensations
    ADD CONSTRAINT employee_compensations_frequency_check
        CHECK (frequency IN ('monthly', 'yearly', 'one_time', 'per_attended_day'));

COMMENT ON COLUMN employee_compensations.frequency IS
    'How often the allowance is paid; per_attended_day uses amount as the daily rate multiplied by days present';
COMMENT ON COLUMN employee_compensations.amount IS
    'Flat amount in cents per period, or the daily rate in cents when frequency = per_attended_day';