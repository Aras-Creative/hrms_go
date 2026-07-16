package entity

import (
	"fmt"
	"strings"
)

type Status string

const (
	StatusActive          Status = "active"
	StatusInactive        Status = "inactive"
	StatusExpiredContract Status = "expired_contract"
	StatusPendingContract Status = "pending_contract"
)

var validStatuses = []Status{
	StatusActive, StatusInactive, StatusExpiredContract, StatusPendingContract,
}

func ParseStatus(s string) (Status, error) {
	st := Status(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validStatuses {
		if st == v {
			return st, nil
		}
	}
	return "", fmt.Errorf("invalid status: %s (must be one of active, inactive, expired_contract, pending_contract)", s)
}

func (s Status) IsValid() bool {
	for _, v := range validStatuses {
		if s == v {
			return true
		}
	}
	return false
}
