package entity

import (
	"fmt"
	"strings"
	"time"
)

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func NewDate(year int, month time.Month, day int) (Date, error) {
	d := Date{Year: year, Month: month, Day: day}
	if !d.IsValid() {
		return Date{}, fmt.Errorf("invalid date: %d-%02d-%02d", year, month, day)
	}
	return d, nil
}

var dateFormats = []string{
	"2006-01-02",
	time.RFC3339,
}

func ParseDate(s string) (Date, error) {
	s = strings.TrimSpace(s)
	for _, f := range dateFormats {
		t, err := time.Parse(f, s)
		if err == nil {
			return Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}, nil
		}
	}
	return Date{}, fmt.Errorf("parse date: unsupported format %q", s)
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) IsZero() bool {
	return d.Year == 0 && d.Month == 0 && d.Day == 0
}

func (d Date) IsValid() bool {
	if d.Year < 1 || d.Year > 9999 {
		return false
	}
	if d.Month < 1 || d.Month > 12 {
		return false
	}
	if d.Day < 1 || d.Day > 31 {
		return false
	}
	_, err := time.Parse("2006-01-02", d.String())
	return err == nil
}

func (d Date) Time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

func (d Date) Age() int {
	now := time.Now()
	age := now.Year() - d.Year
	if now.Month() < d.Month || (now.Month() == d.Month && now.Day() < d.Day) {
		age--
	}
	return age
}
