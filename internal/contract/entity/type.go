package entity

import (
	"fmt"
	"strings"
)

type ContractType string

const (
	ContractTypePKWT  ContractType = "PKWT"
	ContractTypePKWTT ContractType = "PKWTT"
)

var validContractTypes = []ContractType{ContractTypePKWT, ContractTypePKWTT}

func ParseContractType(s string) (ContractType, error) {
	cleaned := ContractType(strings.ToUpper(strings.TrimSpace(s)))
	for _, v := range validContractTypes {
		if cleaned == v {
			return cleaned, nil
		}
	}
	return "", fmt.Errorf("invalid contract type: %s (valid: PKWT, PKWTT)", s)
}
