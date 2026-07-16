ALTER TABLE punches ADD COLUMN date DATE NOT NULL DEFAULT CURRENT_DATE;
CREATE INDEX idx_punches_date ON punches (employee_id, date);
