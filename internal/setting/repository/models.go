package repository

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// SettingModel is the DB representation of app settings.
type SettingModel struct {
	ID               string      `db:"id"`
	Timezone         string      `db:"timezone"`
	CompanyName      string      `db:"company_name"`
	CompanyAddress   string      `db:"company_address"`
	CompanyLogoID    *string     `db:"company_logo_id"`
	WhitelistIPCIDRs StringSlice `db:"whitelist_ip_cidrs"`
	CreatedAt        string      `db:"created_at"`
	UpdatedAt        string      `db:"updated_at"`
}

// StringSlice supports JSONB scan/value for []string.
type StringSlice []string

func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = nil
		return nil
	}
	var raw []byte
	switch v := src.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported type for StringSlice: %T", src)
	}
	return json.Unmarshal(raw, s)
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}
