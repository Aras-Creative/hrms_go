package delivery

import (
	"time"
)

type MemberItem struct {
	ID              string  `json:"id"`
	FullName        string  `json:"full_name"`
	EmployeeNumber  string  `json:"employee_number"`
	ProfilePhotoID  *string `json:"profile_photo_id,omitempty"`
	ProfilePhotoURL string  `json:"profile_photo_url,omitempty"`
}

type DesignationResponse struct {
	ID          string       `json:"id"`
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	Members     []MemberItem `json:"members,omitempty"`
	MemberCount int          `json:"member_count"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
