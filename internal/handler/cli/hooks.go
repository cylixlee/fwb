package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/handler"
)

func (c *Struct) pingCommand() *cobra.Command {
	return &cobra.Command{
		Use:    consts.CommandPing,
		Short:  handler.Translate(consts.KeyCommandPing),
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.hooks.Ping(cmd.Context())
		},
	}
}

func (c *Struct) installCommand() *cobra.Command {
	return &cobra.Command{
		Use:    consts.CommandInstall + " <localPath> <appName>",
		Short:  handler.Translate(consts.KeyCommandInstall),
		Hidden: true,
		Args:   argsExactly(2, consts.KeyUsageInstall),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.hooks.Install(cmd.Context(), args[0])
		},
	}
}

func (c *Struct) uninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:    consts.CommandUninstall + " <appName>",
		Short:  handler.Translate(consts.KeyCommandUninstall),
		Hidden: true,
		Args:   argsExactly(1, consts.KeyUsageUninstall),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.hooks.Uninstall(cmd.Context())
		},
	}
}

func (c *Struct) launchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    consts.CommandLaunch + " <appName> [engineOptions]",
		Short:  handler.Translate(consts.KeyCommandLaunch),
		Hidden: true,
		Args:   argsAtLeast(1, consts.KeyUsageLaunch),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := strings.Join(args[1:], consts.ArgumentSeparator)
			return c.launcher.Launch(cmd.Context(), options, os.Stdout, c.reporter)
		},
	}
	// Flutter expands ${engineOptions} into separate argv entries that start
	// with "--"; stop flag parsing after the app name.
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func (c *Struct) installDeviceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandInstallDevice,
		Short: handler.Translate(consts.KeyCommandInstallDevice),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, enabled, err := c.hooks.Register(cmd.Context(), handler.Translate(consts.KeyDeviceLabel), handler.Translate(consts.KeyDeviceSDKName))
			if err != nil {
				return err
			}
			fmt.Println(handler.Translate(consts.KeyDeviceInstalled, c.cfg.Get(consts.ConfigKeyDeviceID), path))
			if !enabled {
				fmt.Println(handler.Translate(consts.KeyDeviceHintEnable))
			}
			fmt.Println(handler.Translate(consts.KeyDeviceHintNext, c.cfg.Get(consts.ConfigKeyDeviceID)))
			return nil
		},
	}
}

func (c *Struct) uninstallDeviceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandUninstallDevice,
		Short: handler.Translate(consts.KeyCommandUninstallDevice),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := c.hooks.Unregister(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Println(handler.Translate(consts.KeyDeviceRemoved, c.cfg.Get(consts.ConfigKeyDeviceID), path))
			return nil
		},
	}
}
