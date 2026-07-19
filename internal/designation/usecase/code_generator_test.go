package usecase

import (
	"context"
	"testing"

	"hrms/internal/designation/entity"
	"hrms/internal/designation/models"
)

type mockDesignationRepo struct {
	codes map[string]*entity.Designation
}

func newMockRepo() *mockDesignationRepo {
	return &mockDesignationRepo{codes: make(map[string]*entity.Designation)}
}

func (m *mockDesignationRepo) Create(_ context.Context, d *entity.Designation) error {
	m.codes[d.Code] = d
	return nil
}

func (m *mockDesignationRepo) FindByID(_ context.Context, id string) (*entity.Designation, error) {
	for _, d := range m.codes {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, nil
}

func (m *mockDesignationRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.Designation, error) {
	return nil, nil
}

func (m *mockDesignationRepo) FindByCode(_ context.Context, code string) (*entity.Designation, error) {
	if d, ok := m.codes[code]; ok {
		return d, nil
	}
	return nil, nil
}

func (m *mockDesignationRepo) FindAll(_ context.Context) ([]models.DesignationReadModel, error) {
	return nil, nil
}

func (m *mockDesignationRepo) Update(_ context.Context, d *entity.Designation) error {
	return nil
}

func (m *mockDesignationRepo) Delete(_ context.Context, id string) error {
	return nil
}

func TestAcronymFromName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Admin", "ADM"},
		{"Admin KOL", "AK"},
		{"Advertiser", "ADV"},
		{"CRM", "CRM"},
		{"Content Creator", "CC"},
		{"Creative Campaign", "CC"},
		{"Customer Service", "CS"},
		{"E-Commerce Staff", "ECS"},
		{"E-Commerce Supervisor", "ECS"},
		{"Finance Accounting Tax", "FAT"},
		{"HRGA", "HRGA"},
		{"Host Live Streaming", "HLS"},
		{"KOL Specialist", "KS"},
		{"Leader CRM", "LC"},
		{"Leader Customer Service", "LCS"},
		{"Magang", "MAG"},
		{"Software Engineer", "SE"},
		{"Team Youtube", "TY"},
		{"Uploader", "UPL"},
		{"Warehouse Staff", "WS"},
		{"", ""},
	}

	uc := &DesignationUsecase{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.AcronymFromName(tt.name)
			if got != tt.expected {
				t.Errorf("AcronymFromName(%q) = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestPadNextLetter(t *testing.T) {
	uc := &DesignationUsecase{}

	tests := []struct {
		current  string
		name     string
		expected string
	}{
		{"CC", "Creative Campaign", "CCR"},
		{"CC", "Content Creator", "CCO"},
		{"CCR", "Content Creator", "CCRO"},
		{"SE", "Software Engineer", "SEO"},
		{"ADM", "Admin", "ADMI"},
	}

	for _, tt := range tests {
		t.Run(tt.current+"+"+tt.name, func(t *testing.T) {
			got := uc.padNextLetter(tt.current, tt.name)
			if got != tt.expected {
				t.Errorf("padNextLetter(%q, %q) = %q, want %q", tt.current, tt.name, got, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueCode_NoCollision(t *testing.T) {
	repo := newMockRepo()
	uc := NewDesignationUsecase(repo)

	code, err := uc.GenerateUniqueCode(context.Background(), "Software Engineer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "SE" {
		t.Errorf("expected SE, got %s", code)
	}
}

func TestGenerateUniqueCode_WithCollision(t *testing.T) {
	repo := newMockRepo()
	uc := NewDesignationUsecase(repo)

	repo.Create(context.Background(), entity.NewDesignation("Creative Campaign", "CC"))

	code, err := uc.GenerateUniqueCode(context.Background(), "Content Creator")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "CC" {
		t.Errorf("code CC already taken, got duplicate")
	}
	t.Logf("Content Creator resolved to: %s", code)
}

func TestGenerateUniqueCode_MultipleCollisions(t *testing.T) {
	repo := newMockRepo()
	uc := NewDesignationUsecase(repo)

	repo.Create(context.Background(), entity.NewDesignation("Creative Campaign", "CC"))
	repo.Create(context.Background(), entity.NewDesignation("Content Creator", "CCO"))

	code, err := uc.GenerateUniqueCode(context.Background(), "Corporate Communication")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "CC" || code == "CCO" {
		t.Errorf("code %s already taken", code)
	}
	t.Logf("Corporate Communication resolved to: %s", code)
}
