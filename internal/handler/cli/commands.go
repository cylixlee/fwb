package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/handler"
)

func (c *Struct) doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandDoctor,
		Short: handler.Translate(consts.KeyCommandDoctor),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := c.diagnostics.Collect(cmd.Context())
			for _, entry := range report.Entries {
				fmt.Printf(consts.FormatDoctorLine, statusToken(entry.Level), strings.TrimSpace(handler.Translate(entry.Code, entry.Args...)))
			}
			return nil
		},
	}
}

func (c *Struct) cleanCommand() *cobra.Command {
	clean := &cobra.Command{
		Use:   consts.CommandClean,
		Short: handler.Translate(consts.KeyCommandClean),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			withSDK, _ := cmd.Flags().GetBool(consts.FlagSDK)
			removed, err := c.cleaner.Clean(cmd.Context(), withSDK)
			if err != nil {
				return err
			}
			for i, path := range removed {
				if i == 0 {
					fmt.Println(handler.Translate(consts.KeyCleanMirror, path))
				} else {
					fmt.Println(handler.Translate(consts.KeyCleanSDK, path))
				}
			}
			return nil
		},
	}
	clean.Flags().Bool(consts.FlagSDK, false, handler.Translate(consts.KeyCleanFlagSDK))
	return clean
}

func (c *Struct) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandVersion,
		Short: handler.Translate(consts.KeyCommandVersion),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.printVersion()
		},
	}
}

func (c *Struct) helpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   consts.CommandHelp + " [command]",
		Short: handler.Translate(consts.KeyCommandHelp),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := cmd.Root().Find(args)
			if err != nil {
				return errors.New(handler.Translate(consts.KeyUnknownCommand, args[0]))
			}
			return target.Help()
		},
	}
}
