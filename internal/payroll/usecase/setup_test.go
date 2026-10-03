package usecase

import "testing"

func TestOptionalAmount(t *testing.T) {
	tests := []struct {
		name      string
		in        *float64
		wantCents int64
		wantNil   bool
		wantErr   bool
	}{
		{name: "nil stays nil (inherit from type)", in: nil, wantNil: true},
		{name: "major units convert to cents", in: float64Ptr(50_000), wantCents: 5_000_000},
		{name: "zero is allowed", in: float64Ptr(0), wantCents: 0},
		{name: "negative rejected", in: float64Ptr(-1), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optionalAmount(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("optionalAmount() error = %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %d", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected a value, got nil")
			}
			if *got != tt.wantCents {
				t.Errorf("cents = %d, want %d", *got, tt.wantCents)
			}
		})
	}
}

func float64Ptr(v float64) *float64 { return &v }
