DROP INDEX IF EXISTS idx_work_pattern_details_type;

ALTER TABLE work_pattern_details
DROP COLUMN IF EXISTS working_type;
