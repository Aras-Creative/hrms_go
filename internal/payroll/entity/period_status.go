package entity

import "fmt"

type PeriodStatus string

const (
	PeriodStatusDraft     PeriodStatus = "draft"
	PeriodStatusProcessed PeriodStatus = "processed"
	PeriodStatusClosed    PeriodStatus = "closed"
)

var validPeriodStatuses = []PeriodStatus{
	PeriodStatusDraft,
	PeriodStatusProcessed,
	PeriodStatusClosed,
}

var allowedPeriodTransitions = map[PeriodStatus][]PeriodStatus{
	PeriodStatusDraft:     {PeriodStatusProcessed},
	PeriodStatusProcessed: {PeriodStatusClosed, PeriodStatusDraft},
}

func ParsePeriodStatus(s string) (PeriodStatus, error) {
	ps := PeriodStatus(s)
	if !ps.IsValid() {
		return "", fmt.Errorf("invalid period status: %s", s)
	}
	return ps, nil
}

func (s PeriodStatus) IsValid() bool {
	for _, v := range validPeriodStatuses {
		if s == v {
			return true
		}
	}
	return false
}

func (s PeriodStatus) CanTransitionTo(next PeriodStatus) bool {
	allowed, ok := allowedPeriodTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}
