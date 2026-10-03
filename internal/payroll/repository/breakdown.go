package repository

import (
	"fmt"
	"strconv"

	"hrms/internal/payroll/entity"
)

// BuildBreakdowns converts the employee's calculation rows into the display breakdowns
// stored on a payslip.
//
// This lives beside the Calc*Row types because it is the only place that knows how a row
// type maps to its breakdown representation. Both the period processor and the manual
// payslip "recalculate" path go through here, so a re-derived payslip is byte-for-byte
// the same shape as one produced by processing a period.
//
// Components that resolve to zero are dropped: a payslip should not carry a zero-value
// line just because the employee qualifies for nothing.
func BuildBreakdowns(
	comps []CalcCompRow,
	deds []CalcDedRow,
	calcCtx entity.CalcContext,
) ([]entity.CompensationBreakdown, []entity.DeductionBreakdown) {
	compOut := make([]entity.CompensationBreakdown, 0, len(comps))
	for _, c := range comps {
		amount := c.CalculateCents(calcCtx)
		if amount == 0 {
			continue
		}
		compOut = append(compOut, entity.CompensationBreakdown{
			CompensationItemID: c.ID,
			Name:               c.Name,
			Amount:             entity.AmountFromCents(amount).Float(),
		})
	}

	dedOut := make([]entity.DeductionBreakdown, 0, len(deds))
	for _, d := range deds {
		amount := d.CalculateCents(calcCtx)
		if amount == 0 {
			continue
		}
		label := d.Name
		if d.RequiresAttendance() {
			label = fmt.Sprintf("%s (%d hari x %s)",
				d.Name,
				calcCtx.UnpaidAbsentDays,
				FormatIDR(d.PerDayRateCents(calcCtx)),
			)
		}
		dedOut = append(dedOut, entity.DeductionBreakdown{
			DeductionTypeID: d.ID,
			Name:            label,
			Amount:          entity.AmountFromCents(amount).Float(),
		})
	}

	return compOut, dedOut
}

// FormatIDR renders a cents amount the way it appears inside an attendance-based
// deduction label.
func FormatIDR(cents int64) string {
	return fmt.Sprintf("Rp %s", strconv.FormatFloat(entity.AmountFromCents(cents).Float(), 'f', -1, 64))
}
