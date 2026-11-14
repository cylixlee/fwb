package consts

// DoctorLevel classifies doctor findings.
type DoctorLevel int

// Doctor report levels.
const (
	DoctorLevelOK DoctorLevel = iota
	DoctorLevelWarn
	DoctorLevelFail
	DoctorLevelInfo
)
