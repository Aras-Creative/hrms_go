ALTER TABLE work_pattern_details
ADD COLUMN working_type VARCHAR(10) NOT NULL DEFAULT 'fixed';

UPDATE work_pattern_details
SET working_type = 'off'
WHERE start_time IS NULL OR end_time IS NULL;

CREATE INDEX idx_work_pattern_details_type ON work_pattern_details (working_type);
