package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	contractUc "hrms/internal/contract/usecase"
	payrollRepo "hrms/internal/payroll/repository"
)

type CompensationFetcherAdapter struct {
	empCompRepo  payrollRepo.EmployeeCompensationRepository
	compItemRepo payrollRepo.CompensationItemRepository
}

func NewCompensationFetcherAdapter(empCompRepo payrollRepo.EmployeeCompensationRepository, compItemRepo payrollRepo.CompensationItemRepository) *CompensationFetcherAdapter {
	return &CompensationFetcherAdapter{empCompRepo: empCompRepo, compItemRepo: compItemRepo}
}

func (a *CompensationFetcherAdapter) FindByEmployeeID(ctx context.Context, employeeID string) ([]contractUc.CompensationItem, error) {
	empComps, err := a.empCompRepo.FindByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("find employee compensations: %w", err)
	}

	type compEntry struct {
		itemID string
		amount float64
	}
	entries := make(map[string]compEntry)
	for _, ec := range empComps {
		if ec.EndDate == nil {
			entries[ec.CompensationItemID] = compEntry{
				itemID: ec.CompensationItemID,
				amount: ec.Amount.Float(),
			}
		}
	}

	items := make([]contractUc.CompensationItem, 0, len(entries))
	for _, entry := range entries {
		ci, err := a.compItemRepo.FindByID(ctx, entry.itemID)
		if err != nil || ci == nil {
			continue
		}
		items = append(items, contractUc.CompensationItem{
			Name:   ci.Name,
			Amount: formatRupiah(entry.amount),
		})
	}
	slog.Debug("compensations fetched", "employee_id", employeeID, "count", len(items), "total_emp_comps", len(empComps), "active_entries", len(entries))
	return items, nil
}

func formatRupiah(amount float64) string {
	n := int64(amount)
	remain := int64((amount - float64(n)) * 100)

	digits := fmt.Sprintf("%d", n)
	var groups []string
	for len(digits) > 3 {
		groups = append(groups, digits[len(digits)-3:])
		digits = digits[:len(digits)-3]
	}
	groups = append(groups, digits)
	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}
	result := "Rp " + strings.Join(groups, ".")
	if remain > 0 {
		result += fmt.Sprintf(",%02d", remain)
	}
	return result
}

var _ contractUc.CompensationFetcher = (*CompensationFetcherAdapter)(nil)
