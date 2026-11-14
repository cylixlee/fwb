package domain

import "github.com/cylixlee/fwb/internal/consts"

// DoctorEntry is one diagnostic line. Code is an i18n key; Args fill it in.
type DoctorEntry struct {
	Level consts.DoctorLevel
	Code  string
	Args  []any
}

// DoctorReport is the result of a doctor run.
type DoctorReport struct {
	Entries []DoctorEntry
}

// Add appends an entry.
func (r *DoctorReport) Add(level consts.DoctorLevel, code string, args ...any) {
	r.Entries = append(r.Entries, DoctorEntry{Level: level, Code: code, Args: args})
}
