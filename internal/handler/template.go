package handler

import (
	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed templates/usage.tmpl
var usageTemplate string

//go:embed templates/help.tmpl
var helpTemplate string

// ApplyTemplates installs the localized usage and help templates, recursively.
func ApplyTemplates(cmd *cobra.Command) {
	cmd.SetUsageTemplate(usageTemplate)
	cmd.SetHelpTemplate(helpTemplate)
	for _, child := range cmd.Commands() {
		ApplyTemplates(child)
	}
}
