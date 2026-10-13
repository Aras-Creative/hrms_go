package middleware

import (
	"context"
	"errors"
	"testing"

	"hrms/internal/pkg/iputil"
)

type stubEmployeeResolver struct {
	employeeID string
	err        error
}

func (s stubEmployeeResolver) FindByUserID(_ context.Context, _ string) (string, string, error) {
	return s.employeeID, "", s.err
}

func TestIPAllowed(t *testing.T) {
	iputil.SetAllowedCIDRs([]string{"10.0.0.0/8"})
	iputil.SetBypassEmployeeIDs([]string{"emp-bypass"})
	defer iputil.SetAllowedCIDRs(nil)
	defer iputil.SetBypassEmployeeIDs(nil)

	tests := []struct {
		name     string
		ip       string
		userID   string
		resolver EmployeeResolver
		want     bool
	}{
		{"allowed ip", "10.1.2.3", "user-1", nil, true},
		{"unlisted ip no user", "8.8.8.8", "", stubEmployeeResolver{employeeID: "emp-bypass"}, false},
		{"unlisted ip no resolver", "8.8.8.8", "user-1", nil, false},
		{"unlisted ip bypassed employee", "8.8.8.8", "user-1", stubEmployeeResolver{employeeID: "emp-bypass"}, true},
		{"unlisted ip non-bypassed employee", "8.8.8.8", "user-1", stubEmployeeResolver{employeeID: "emp-other"}, false},
		{"resolver error", "8.8.8.8", "user-1", stubEmployeeResolver{err: errors.New("boom")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ipAllowed(context.Background(), tt.ip, tt.userID, tt.resolver); got != tt.want {
				t.Errorf("ipAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
