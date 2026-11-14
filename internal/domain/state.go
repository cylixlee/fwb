package domain

import "time"

// BuildState is the persisted result of the last Windows build.
type BuildState struct {
	Fingerprint    string
	EngineRevision string
	BuildDir       string
	Exe            string
	BuiltAt        time.Time
}
