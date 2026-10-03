-- Drops the descriptive figures. Re-entering them is a manual act on the slip, but the
-- audit log keeps the history of what they used to say.
ALTER TABLE pay_slips
    DROP COLUMN IF EXISTS income_inputs;