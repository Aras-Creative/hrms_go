package iputil

import "testing"

func TestIsAllowed(t *testing.T) {
	SetAllowedCIDRs([]string{"10.0.0.0/8", "192.168.1.10"})
	defer SetAllowedCIDRs(nil)

	tests := []struct {
		ip      string
		allowed bool
	}{
		{"10.1.2.3", true},
		{"10.1.2.3:5678", true},
		{"192.168.1.10", true},
		{"192.168.1.11", false},
		{"8.8.8.8", false},
		{"not-an-ip", false},
	}

	for _, tt := range tests {
		if got := IsAllowed(tt.ip); got != tt.allowed {
			t.Errorf("IsAllowed(%q) = %v, want %v", tt.ip, got, tt.allowed)
		}
	}
}

func TestIsAllowedEmptyListAllowsAll(t *testing.T) {
	SetAllowedCIDRs(nil)
	if !IsAllowed("8.8.8.8") {
		t.Fatal("expected empty allow-list to allow every IP")
	}
}

func TestBypassEmployeeIDs(t *testing.T) {
	SetBypassEmployeeIDs([]string{"emp-1", " emp-2 ", ""})
	defer SetBypassEmployeeIDs(nil)

	if !IsEmployeeBypassed("emp-1") {
		t.Error("expected emp-1 to be bypassed")
	}
	if !IsEmployeeBypassed("emp-2") {
		t.Error("expected trimmed emp-2 to be bypassed")
	}
	if IsEmployeeBypassed("emp-3") {
		t.Error("did not expect emp-3 to be bypassed")
	}
	if IsEmployeeBypassed("") {
		t.Error("did not expect empty employee ID to be bypassed")
	}

	SetBypassEmployeeIDs(nil)
	if IsEmployeeBypassed("emp-1") {
		t.Error("expected bypass list to be cleared")
	}
}
