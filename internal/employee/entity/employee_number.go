package entity

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var employeeNumberRegex = regexp.MustCompile(`^(\d{4})([A-Z]{2,10})(\d{2})$`)

type EmployeeNumber struct {
	value string
}

func NewEmployeeNumber(value string) (EmployeeNumber, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return EmployeeNumber{}, fmt.Errorf("employee_number is required")
	}
	if !employeeNumberRegex.MatchString(v) {
		return EmployeeNumber{}, fmt.Errorf("employee_number must match format: 4 digits + 3-10 uppercase letters + 2 digits (e.g. 1001SWE01)")
	}
	return EmployeeNumber{value: v}, nil
}

func FromString(value string) EmployeeNumber {
	return EmployeeNumber{value: value}
}

func (n EmployeeNumber) String() string {
	return n.value
}

func (n EmployeeNumber) IsEmpty() bool {
	return n.value == ""
}

func (n EmployeeNumber) Code() string {
	matches := employeeNumberRegex.FindStringSubmatch(n.value)
	if matches == nil {
		return ""
	}
	return matches[2]
}

func (n EmployeeNumber) Sequence() int {
	matches := employeeNumberRegex.FindStringSubmatch(n.value)
	if matches == nil {
		return 0
	}
	seq, _ := strconv.Atoi(matches[3])
	return seq
}
