package entity

import (
	"fmt"
	"strings"
)

type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"
	LeaveStatusApproved  LeaveStatus = "approved"
	LeaveStatusRejected  LeaveStatus = "rejected"
	LeaveStatusCancelled LeaveStatus = "cancelled"
)

var validLeaveStatuses = []LeaveStatus{
	LeaveStatusPending, LeaveStatusApproved, LeaveStatusRejected, LeaveStatusCancelled,
}

var allowedTransitions = map[LeaveStatus][]LeaveStatus{
	LeaveStatusPending:  {LeaveStatusApproved, LeaveStatusRejected, LeaveStatusCancelled},
	LeaveStatusApproved: {LeaveStatusCancelled},
	LeaveStatusRejected: {},
	LeaveStatusCancelled: {},
}

func ParseLeaveStatus(s string) (LeaveStatus, error) {
	st := LeaveStatus(strings.ToLower(strings.TrimSpace(s)))
	for _, v := range validLeaveStatuses {
		if st == v {
			return st, nil
		}
	}
	return "", fmt.Errorf("invalid leave status: %s", s)
}

func (s LeaveStatus) CanTransitionTo(next LeaveStatus) bool {
	allowed, ok := allowedTransitions[s]
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
