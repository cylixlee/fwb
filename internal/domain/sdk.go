package domain

import (
	"path/filepath"

	"github.com/cylixlee/fwb/internal/consts"
)

// WindowsSDK is a downloaded (or configured) Windows Flutter SDK.
type WindowsSDK struct {
	Version     FlutterVersion
	Root        string // cache root, e.g. .../sdk/<version>-<channel>
	FlutterRoot string // the directory that contains bin/flutter.bat
}

// FlutterBat is the Windows launcher inside the SDK.
func (s WindowsSDK) FlutterBat() string {
	return filepath.Join(s.FlutterRoot, consts.FlutterBinDirectory, consts.FlutterBatFileName)
}
