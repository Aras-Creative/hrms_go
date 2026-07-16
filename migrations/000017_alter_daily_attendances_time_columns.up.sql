ALTER TABLE daily_attendances
    ALTER COLUMN expected_start_time TYPE VARCHAR(5) USING to_char(expected_start_time, 'HH24:MI'),
    ALTER COLUMN expected_end_time TYPE VARCHAR(5) USING to_char(expected_end_time, 'HH24:MI');
