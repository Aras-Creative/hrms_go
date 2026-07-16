package entity

import (
	"fmt"
	"strings"
)

type ContractStatus string

const (
	ContractStatusDraft      ContractStatus = "draft"
	ContractStatusSent       ContractStatus = "sent"
	ContractStatusActive     ContractStatus = "active"
	ContractStatusExpired    ContractStatus = "expired"
	ContractStatusTerminated ContractStatus = "terminated"
)

var validContractStatuses = []ContractStatus{
	ContractStatusDraft,
	ContractStatusSent,
	ContractStatusActive,
	ContractStatusExpired,
	ContractStatusTerminated,
}

func ParseContractStatus(s string) (ContractStatus, error) {
	cleaned := ContractStatus(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validContractStatuses {
		if cleaned == v {
			return cleaned, nil
		}
	}
	return "", fmt.Errorf("invalid contract status: %s (valid: draft, sent, active, expired, terminated)", s)
}
