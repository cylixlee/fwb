package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/cylixlee/fwb/internal/handler"
)

func argsExactly(n int, usageKey string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return errors.New(handler.Translate(usageKey))
		}
		return nil
	}
}

func argsAtLeast(n int, usageKey string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return errors.New(handler.Translate(usageKey))
		}
		return nil
	}
}
