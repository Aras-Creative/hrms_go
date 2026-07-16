ALTER TABLE payroll_periods DROP CONSTRAINT payroll_periods_status_check;
ALTER TABLE payroll_periods ADD CONSTRAINT payroll_periods_status_check CHECK (status IN ('draft', 'closed'));
