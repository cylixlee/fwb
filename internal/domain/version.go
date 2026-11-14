package domain

import "github.com/cylixlee/fwb/internal/consts"

// Version identifies a Flutter SDK.
type FlutterVersion struct {
	Framework         string
	Channel           string
	FrameworkRevision string
	EngineRevision    string
	DartSDK           string
}

// ShortEngine returns the first 12 characters of the engine revision.
func (v FlutterVersion) ShortEngine() string {
	if len(v.EngineRevision) > consts.ShortEngineLength {
		return v.EngineRevision[:consts.ShortEngineLength]
	}
	return v.EngineRevision
}

// SameEngine reports whether both versions are backed by the same engine.
func (v FlutterVersion) SameEngine(other FlutterVersion) bool {
	return v.EngineRevision != "" && v.EngineRevision == other.EngineRevision
}
