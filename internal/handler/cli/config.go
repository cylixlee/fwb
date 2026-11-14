package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/handler"
)

func (c *Struct) configCommand() *cobra.Command {
	get := &cobra.Command{
		Use:   consts.ConfigCommandGet + " <key>",
		Short: handler.Translate(consts.KeyCommandConfigGet),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if all, _ := cmd.Flags().GetBool(consts.FlagAll); all {
				c.printConfig()
				return nil
			}
			if len(args) == 0 {
				return errors.New(handler.Translate(consts.KeyUsageConfigGet))
			}
			fmt.Println(c.cfg.Get(args[0]))
			return nil
		},
	}
	// Same --all flag as `config list`.
	get.Flags().BoolP(consts.FlagAll, consts.ShorthandAll, false, handler.Translate(consts.KeyFlagAll))

	set := &cobra.Command{
		Use:   consts.ConfigCommandSet + " <key> <value>",
		Short: handler.Translate(consts.KeyCommandConfigSet),
		Args:  argsAtLeast(2, consts.KeyUsageConfigSet),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], strings.Join(args[1:], consts.ArgumentSeparator)
			if err := c.cfg.Set(key, value); err != nil {
				return err
			}
			fmt.Println(handler.Translate(consts.KeyConfigSet, key, c.cfg.Get(key)))
			return nil
		},
	}

	unset := &cobra.Command{
		Use:   consts.ConfigCommandUnset + " <key>",
		Short: handler.Translate(consts.KeyCommandConfigUnset),
		Args:  argsExactly(1, consts.KeyUsageConfigUnset),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.cfg.Unset(args[0])
		},
	}

	list := &cobra.Command{
		Use:   consts.ConfigCommandList,
		Short: handler.Translate(consts.KeyCommandConfigList),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.printConfig()
			return nil
		},
	}

	config := &cobra.Command{
		Use:   consts.CommandConfig,
		Short: handler.Translate(consts.KeyCommandConfig),
	}
	config.AddCommand(get, set, unset, list)
	return config
}

func (c *Struct) printConfig() {
	all := c.cfg.All()
	for _, key := range c.cfg.Keys() {
		fmt.Printf(consts.FormatConfigLine, key, all[key], c.cfg.Source(key))
	}
}
