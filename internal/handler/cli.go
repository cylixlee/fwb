package handler

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/cylixlee/fwb/internal/consts"
	"os"
	"strings"

	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/srv"
)

// Version is overridden at build time with -ldflags.
var BuildVersion = consts.DefaultBuildVersion

// CLI is the command handler.
type CLI struct {
	svc *srv.Bridge
}

// New returns the CLI handler.
func NewCLI(svc *srv.Bridge) *CLI { return &CLI{svc: svc} }

// Execute runs one command and returns the process exit code.
func (h *CLI) Execute(ctx context.Context, args []string) int {
	if lang := h.svc.Language(); lang != "" {
		SetLanguage(lang)
	}
	if len(args) == 0 {
		fmt.Println(Translate(consts.KeyHelpText))
		return 0
	}
	cmd, rest := args[0], args[1:]
	var err error
	switch cmd {
	case consts.CommandVersion, consts.FlagVersion, consts.FlagVersionShort:
		fmt.Printf(consts.FormatVersion, BuildVersion, current)
	case consts.CommandHelp, consts.FlagHelp, consts.FlagHelpShort:
		fmt.Println(Translate(consts.KeyHelpText))
	case consts.CommandPing:
		err = h.svc.Ping(ctx)
	case consts.CommandUninstall:
		if len(rest) < 1 {
			err = errors.New(Translate(consts.KeyUsageUninstall))
			break
		}
		err = h.svc.Uninstall(ctx)
	case consts.CommandInstall:
		if len(rest) < 2 {
			err = errors.New(Translate(consts.KeyUsageInstall))
			break
		}
		err = h.svc.Install(ctx, rest[0])
	case consts.CommandLaunch:
		if len(rest) < 1 {
			err = errors.New(Translate(consts.KeyUsageLaunch))
			break
		}
		err = h.svc.Launch(ctx, strings.Join(rest[1:], consts.ArgumentSeparator), os.Stdout, reporter{})
	case consts.CommandConfig:
		err = h.config(rest)
	case consts.CommandInstallDevice:
		err = h.installDevice(ctx)
	case consts.CommandUninstallDevice:
		err = h.uninstallDevice(ctx)
	case consts.CommandDoctor:
		h.doctor(ctx)
	case consts.CommandClean:
		err = h.clean(ctx, rest)
	default:
		err = fmt.Errorf(Translate(consts.KeyUnknownCommand), cmd)
	}
	if err != nil {
		h.fail(err)
		return 1
	}
	return 0
}

func (h *CLI) config(args []string) error {
	if len(args) == 0 {
		return errors.New(Translate(consts.KeyUsageConfig))
	}
	switch args[0] {
	case consts.ConfigCommandSet:
		if len(args) < 3 {
			return errors.New(Translate(consts.KeyUsageConfigSet))
		}
		key, value := args[1], strings.Join(args[2:], consts.ArgumentSeparator)
		if err := h.svc.ConfigSet(key, value); err != nil {
			return err
		}
		fmt.Println(Translate(consts.KeyConfigSet, key, h.svc.ConfigGet(key)))
	case consts.ConfigCommandGet:
		if len(args) < 2 {
			return errors.New(Translate(consts.KeyUsageConfigGet))
		}
		if args[1] == consts.FlagAll || args[1] == consts.FlagAllShort {
			h.listConfig()
			return nil
		}
		fmt.Println(h.svc.ConfigGet(args[1]))
	case consts.ConfigCommandUnset:
		if len(args) < 2 {
			return errors.New(Translate(consts.KeyUsageConfigUnset))
		}
		return h.svc.ConfigUnset(args[1])
	case consts.ConfigCommandList, consts.FlagAll, consts.FlagAllShort:
		h.listConfig()
	default:
		return fmt.Errorf(Translate(consts.KeyUnknownConfigCommand), args[0])
	}
	return nil
}

func (h *CLI) listConfig() {
	all := h.svc.ConfigAll()
	for _, key := range h.svc.ConfigKeys() {
		fmt.Printf(consts.FormatConfigLine, key, all[key], h.svc.ConfigSource(key))
	}
}

func (h *CLI) installDevice(ctx context.Context) error {
	path, enabled, err := h.svc.RegisterDevice(ctx, Translate(consts.KeyDeviceLabel), Translate(consts.KeyDeviceSDKName))
	if err != nil {
		return err
	}
	fmt.Println(Translate(consts.KeyDeviceInstalled, h.svc.DeviceID(), path))
	if !enabled {
		fmt.Println(Translate(consts.KeyDeviceHintEnable))
	}
	fmt.Println(Translate(consts.KeyDeviceHintNext, h.svc.DeviceID()))
	return nil
}

func (h *CLI) uninstallDevice(ctx context.Context) error {
	path, err := h.svc.UnregisterDevice(ctx)
	if err != nil {
		return err
	}
	fmt.Println(Translate(consts.KeyDeviceRemoved, h.svc.DeviceID(), path))
	return nil
}

func (h *CLI) doctor(ctx context.Context) {
	report := h.svc.Doctor(ctx)
	for _, e := range report.Entries {
		fmt.Printf(consts.FormatDoctorLine, statusToken(e.Level), strings.TrimSpace(Translate(e.Code, e.Args...)))
	}
}

func (h *CLI) clean(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet(consts.CommandClean, flag.ContinueOnError)
	fs.Usage = func() {}
	withSDK := fs.Bool(consts.FlagSDK, false, consts.EmptyUsage)
	if err := fs.Parse(args); err != nil {
		return err
	}
	removed, err := h.svc.Clean(ctx, *withSDK)
	if err != nil {
		return err
	}
	for i, path := range removed {
		if i == 0 {
			fmt.Println(Translate(consts.KeyCleanMirror, path))
		} else {
			fmt.Println(Translate(consts.KeyCleanSDK, path))
		}
	}
	return nil
}

func (h *CLI) fail(err error) {
	fmt.Fprintf(os.Stderr, consts.FormatError, message(err))
}

func message(err error) string {
	if domainErr, ok := errors.AsType[*domain.Error](err); ok {
		msg := Translate(consts.ErrorKeyPrefix+string(domainErr.Code), domainErr.Args...)
		if domainErr.Cause != nil {
			return msg + consts.ErrorCauseSeparator + message(domainErr.Cause)
		}
		return msg
	}
	return err.Error()
}

func statusToken(level consts.DoctorLevel) string {
	switch level {
	case consts.DoctorLevelOK:
		return Translate(consts.KeyDoctorStatusOK)
	case consts.DoctorLevelWarn:
		return Translate(consts.KeyDoctorStatusWarn)
	case consts.DoctorLevelFail:
		return Translate(consts.KeyDoctorStatusFail)
	default:
		return Translate(consts.KeyDoctorStatusInfo)
	}
}
