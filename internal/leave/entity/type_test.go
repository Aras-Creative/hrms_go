package entity

import (
	"testing"
)

func TestLeaveType(t *testing.T) {
	lt := NewLeaveType("Cuti Tahunan", 12, true, false, false, true)

	if lt.Name != "Cuti Tahunan" {
		t.Errorf("expected name 'Cuti Tahunan', got '%s'", lt.Name)
	}
	if lt.DefaultDays != 12 {
		t.Errorf("expected default days 12, got %d", lt.DefaultDays)
	}
	if !lt.IsPaid {
		t.Errorf("expected is_paid true")
	}
	if !lt.IncludeInAttendanceAllowance {
		t.Errorf("expected include_in_attendance_allowance true")
	}

	lt.SetIncludeInAttendanceAllowance(false)
	if lt.IncludeInAttendanceAllowance {
		t.Errorf("expected include_in_attendance_allowance false after update")
	}
}
