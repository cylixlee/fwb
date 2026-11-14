package cli

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"golang.org/x/mod/semver"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/handler"
)

func (c *Struct) printVersion() error {
	fmt.Printf(consts.FormatVersion, moduleVersion(), handler.Current())
	return nil
}

func (c *Struct) fail(err error) {
	fmt.Fprintf(os.Stderr, consts.FormatError, errorMessage(err))
}

func errorMessage(err error) string {
	if domainErr, ok := errors.AsType[*domain.Error](err); ok {
		msg := handler.Translate(consts.ErrorKeyPrefix+string(domainErr.Code), domainErr.Args...)
		if domainErr.Cause != nil {
			return msg + consts.ErrorCauseSeparator + errorMessage(domainErr.Cause)
		}
		return msg
	}
	return err.Error()
}

func statusToken(level consts.DoctorLevel) string {
	switch level {
	case consts.DoctorLevelOK:
		return handler.Translate(consts.KeyDoctorStatusOK)
	case consts.DoctorLevelWarn:
		return handler.Translate(consts.KeyDoctorStatusWarn)
	case consts.DoctorLevelFail:
		return handler.Translate(consts.KeyDoctorStatusFail)
	default:
		return handler.Translate(consts.KeyDoctorStatusInfo)
	}
}

// moduleVersion returns a semver version. For `go install module@version` this
// is the module version recorded by the toolchain; for working-tree builds it
// falls back to v0.0.0-dev (plus the VCS revision when available).
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		if semver.IsValid(info.Main.Version) {
			return info.Main.Version
		}
		if revision := buildSetting(info, consts.BuildSettingVCSRevision); revision != "" {
			if len(revision) > consts.VCSRevisionShortLength {
				revision = revision[:consts.VCSRevisionShortLength]
			}
			return consts.DefaultBuildVersion + "+" + revision
		}
	}
	return consts.DefaultBuildVersion
}

func buildSetting(info *debug.BuildInfo, key string) string {
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}
