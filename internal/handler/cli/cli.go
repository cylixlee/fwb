package cli

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/internal/handler"
	"github.com/cylixlee/fwb/internal/srv"
)

// Struct is the cobra command tree.
type Struct struct {
	cfg         domain.ConfigStore
	hooks       *srv.DeviceHooks
	launcher    *srv.Launcher
	cleaner     *srv.Cleaner
	diagnostics *srv.Diagnostics
	root        *cobra.Command
	reporter    domain.EventReporter
}

// New builds the command tree using the configured language.
func New(
	cfg domain.ConfigStore,
	hooks *srv.DeviceHooks,
	launcher *srv.Launcher,
	cleaner *srv.Cleaner,
	diagnostics *srv.Diagnostics,
) *Struct {
	if lang := cfg.Get(consts.ConfigKeyLanguage); lang != "" {
		handler.SetLanguage(lang)
	}
	cobra.AddTemplateFunc("T", handler.Translate)
	command := &Struct{
		cfg:         cfg,
		hooks:       hooks,
		launcher:    launcher,
		cleaner:     cleaner,
		diagnostics: diagnostics,
		reporter:    handler.NewReporter(),
	}
	command.root = command.newRoot()
	return command
}

// Execute runs one command and returns the process exit code.
func (c *Struct) Execute(ctx context.Context, args []string) int {
	if unknown := c.unknownCommand(args); unknown != "" {
		c.fail(errors.New(handler.Translate(consts.KeyUnknownCommand, unknown)))
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

func (c *Struct) newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           consts.CommandName,
		Short:         handler.Translate(consts.KeyCommandRoot),
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
	root.PersistentFlags().BoolP(consts.FlagHelp, consts.ShorthandHelp, false, handler.Translate(consts.KeyFlagHelp))
	root.PersistentFlags().BoolP(consts.FlagVersion, consts.ShorthandVersion, false, handler.Translate(consts.KeyFlagVersion))
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
	handler.ApplyTemplates(root)
	return root
}

func (c *Struct) unknownCommand(args []string) string {
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
		next, _, err := cmd.Find([]string{arg})
		if err != nil || next == cmd {
			return arg
		}
		cmd = next
	}
	return ""
}
