ALTER TABLE daily_attendances
    ALTER COLUMN expected_start_time TYPE TIME USING expected_start_time::time,
    ALTER COLUMN expected_end_time TYPE TIME USING expected_end_time::time;
