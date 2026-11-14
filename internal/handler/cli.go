package handler

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/srv"
)

//go:embed templates/usage.tmpl
var usageTemplate string

//go:embed templates/help.tmpl
var helpTemplate string

// CLI is the cobra command handler.
type CLI struct {
	bridge   *srv.Bridge
	root     *cobra.Command
	reporter *reporter
}

// NewCLI builds the command tree using the configured language.
func NewCLI(bridge *srv.Bridge) *CLI {
	if lang := bridge.Language(); lang != "" {
		SetLanguage(lang)
	}
	cobra.AddTemplateFunc("T", Translate)
	cli := &CLI{bridge: bridge, reporter: newStdoutReporter()}
	cli.root = cli.newRoot()
	return cli
}

// Execute runs one command and returns the process exit code.
func (c *CLI) Execute(ctx context.Context, args []string) int {
	if unknown := c.unknownCommand(args); unknown != "" {
		c.fail(errors.New(Translate(consts.KeyUnknownCommand, unknown)))
		return consts.ExitFailure
	}
	c.root.SetArgs(args)
	c.root.SetOut(os.Stdout)
	c.root.SetErr(os.Stderr)
	if err := c.root.ExecuteContext(ctx); err != nil {
		c.fail(err)
		return consts.ExitFailure
	}
	return consts.ExitSuccess
}

func (c *CLI) newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           consts.CommandName,
		Short:         Translate(consts.KeyCommandRoot),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if version, _ := cmd.Flags().GetBool(consts.FlagVersion); version {
				return c.printVersion()
			}
			return cmd.Help()
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().BoolP(consts.FlagHelp, consts.ShorthandHelp, false, Translate(consts.KeyFlagHelp))
	root.PersistentFlags().BoolP(consts.FlagVersion, consts.ShorthandVersion, false, Translate(consts.KeyFlagVersion))
	root.SetHelpCommand(c.helpCommand())
	root.AddCommand(
		c.pingCommand(),
		c.installCommand(),
		c.uninstallCommand(),
		c.launchCommand(),
		c.configCommand(),
		c.installDeviceCommand(),
		c.uninstallDeviceCommand(),
		c.doctorCommand(),
		c.cleanCommand(),
		c.versionCommand(),
	)
	applyTemplates(root)
	return root
}

func (c *CLI) pingCommand() *cobra.Command {
	return &cobra.Command{
		Use:    consts.CommandPing,
		Short:  Translate(consts.KeyCommandPing),
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.bridge.Ping(cmd.Context())
		},
	}
}

func (c *CLI) installCommand() *cobra.Command {
	return &cobra.Command{
		Use:    consts.CommandInstall + " <localPath> <appName>",
		Short:  Translate(consts.KeyCommandInstall),
		Hidden: true,
		Args:   argsExactly(2, consts.KeyUsageInstall),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.bridge.Install(cmd.Context(), args[0])
		},
	}
}

func (c *CLI) uninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:    consts.CommandUninstall + " <appName>",
		Short:  Translate(consts.KeyCommandUninstall),
		Hidden: true,
		Args:   argsExactly(1, consts.KeyUsageUninstall),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.bridge.Uninstall(cmd.Context())
		},
	}
}

func (c *CLI) launchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    consts.CommandLaunch + " <appName> [engineOptions]",
		Short:  Translate(consts.KeyCommandLaunch),
		Hidden: true,
		Args:   argsAtLeast(1, consts.KeyUsageLaunch),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := strings.Join(args[1:], consts.ArgumentSeparator)
			return c.bridge.Launch(cmd.Context(), options, os.Stdout, c.reporter)
		},
	}
	// Flutter expands ${engineOptions} into separate argv entries that start
	// with "--"; stop flag parsing after the app name.
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func (c *CLI) configCommand() *cobra.Command {
	config := &cobra.Command{
		Use:   consts.CommandConfig,
		Short: Translate(consts.KeyCommandConfig),
	}
	config.AddCommand(
		&cobra.Command{
			Use:   consts.ConfigCommandSet + " <key> <value>",
			Short: Translate(consts.KeyCommandConfigSet),
			Args:  argsAtLeast(2, consts.KeyUsageConfigSet),
			RunE: func(cmd *cobra.Command, args []string) error {
				key, value := args[0], strings.Join(args[1:], consts.ArgumentSeparator)
				if err := c.bridge.ConfigSet(key, value); err != nil {
					return err
				}
				fmt.Println(Translate(consts.KeyConfigSet, key, c.bridge.ConfigGet(key)))
				return nil
			},
		},
		&cobra.Command{
			Use:   consts.ConfigCommandGet + " <key>",
			Short: Translate(consts.KeyCommandConfigGet),
			Args:  cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				if all, _ := cmd.Flags().GetBool(consts.FlagAll); all {
					c.printConfig()
					return nil
				}
				if len(args) == 0 {
					return errors.New(Translate(consts.KeyUsageConfigGet))
				}
				fmt.Println(c.bridge.ConfigGet(args[0]))
				return nil
			},
		},
		&cobra.Command{
			Use:   consts.ConfigCommandUnset + " <key>",
			Short: Translate(consts.KeyCommandConfigUnset),
			Args:  argsExactly(1, consts.KeyUsageConfigUnset),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.bridge.ConfigUnset(args[0])
			},
		},
		&cobra.Command{
			Use:   consts.ConfigCommandList,
			Short: Translate(consts.KeyCommandConfigList),
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				c.printConfig()
				return nil
			},
		},
	)
	// Same --all flag as `config get --all`.
	if get := findCommand(config, consts.ConfigCommandGet); get != nil {
		get.Flags().BoolP(consts.FlagAll, consts.ShorthandAll, false, Translate(consts.KeyFlagAll))
	}
	return config
}

func (c *CLI) installDeviceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandInstallDevice,
		Short: Translate(consts.KeyCommandInstallDevice),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, enabled, err := c.bridge.RegisterDevice(cmd.Context(), Translate(consts.KeyDeviceLabel), Translate(consts.KeyDeviceSDKName))
			if err != nil {
				return err
			}
			fmt.Println(Translate(consts.KeyDeviceInstalled, c.bridge.DeviceID(), path))
			if !enabled {
				fmt.Println(Translate(consts.KeyDeviceHintEnable))
			}
			fmt.Println(Translate(consts.KeyDeviceHintNext, c.bridge.DeviceID()))
			return nil
		},
	}
}

func (c *CLI) uninstallDeviceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandUninstallDevice,
		Short: Translate(consts.KeyCommandUninstallDevice),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := c.bridge.UnregisterDevice(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Println(Translate(consts.KeyDeviceRemoved, c.bridge.DeviceID(), path))
			return nil
		},
	}
}

func (c *CLI) doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandDoctor,
		Short: Translate(consts.KeyCommandDoctor),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := c.bridge.Doctor(cmd.Context())
			for _, entry := range report.Entries {
				fmt.Printf(consts.FormatDoctorLine, statusToken(entry.Level), strings.TrimSpace(Translate(entry.Code, entry.Args...)))
			}
			return nil
		},
	}
}

func (c *CLI) cleanCommand() *cobra.Command {
	clean := &cobra.Command{
		Use:   consts.CommandClean,
		Short: Translate(consts.KeyCommandClean),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			withSDK, _ := cmd.Flags().GetBool(consts.FlagSDK)
			removed, err := c.bridge.Clean(cmd.Context(), withSDK)
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
		},
	}
	clean.Flags().Bool(consts.FlagSDK, false, Translate(consts.KeyCleanFlagSDK))
	return clean
}

func (c *CLI) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandVersion,
		Short: Translate(consts.KeyCommandVersion),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.printVersion()
		},
	}
}

func (c *CLI) helpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandHelp + " [command]",
		Short: Translate(consts.KeyCommandHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := cmd.Root().Find(args)
			if err != nil {
				return errors.New(Translate(consts.KeyUnknownCommand, args[0]))
			}
			return target.Help()
		},
	}
}

func (c *CLI) printVersion() error {
	fmt.Printf(consts.FormatVersion, moduleVersion(), Current())
	return nil
}

func (c *CLI) printConfig() {
	all := c.bridge.ConfigAll()
	for _, key := range c.bridge.ConfigKeys() {
		fmt.Printf(consts.FormatConfigLine, key, all[key], c.bridge.ConfigSource(key))
	}
}

func (c *CLI) fail(err error) {
	fmt.Fprintf(os.Stderr, consts.FormatError, errorMessage(err))
}

func errorMessage(err error) string {
	if domainErr, ok := errors.AsType[*domain.Error](err); ok {
		msg := Translate(consts.ErrorKeyPrefix+string(domainErr.Code), domainErr.Args...)
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
		return Translate(consts.KeyDoctorStatusOK)
	case consts.DoctorLevelWarn:
		return Translate(consts.KeyDoctorStatusWarn)
	case consts.DoctorLevelFail:
		return Translate(consts.KeyDoctorStatusFail)
	default:
		return Translate(consts.KeyDoctorStatusInfo)
	}
}

func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func (c *CLI) unknownCommand(args []string) string {
	cmd := c.root
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if arg == consts.CommandHelp {
			return ""
		}
		if len(cmd.Commands()) == 0 {
			return ""
		}
		next := findCommand(cmd, arg)
		if next == nil {
			return arg
		}
		cmd = next
	}
	return ""
}

func argsExactly(n int, usageKey string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return errors.New(Translate(usageKey))
		}
		return nil
	}
}

func argsAtLeast(n int, usageKey string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return errors.New(Translate(usageKey))
		}
		return nil
	}
}

func applyTemplates(cmd *cobra.Command) {
	cmd.SetUsageTemplate(usageTemplate)
	cmd.SetHelpTemplate(helpTemplate)
	for _, child := range cmd.Commands() {
		applyTemplates(child)
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
