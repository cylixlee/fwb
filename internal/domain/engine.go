package domain

import (
	"strings"

	"github.com/mattn/go-shellwords"

	"github.com/cylixlee/fwb/internal/consts"
)

// ParseEngineSwitches splits the flutter tool's `${engineOptions}` value into
// individual engine switches. The value arrives as one argv element because the
// custom-device hook is exec'd without a shell. Quoted values keep spaces.
func ParseEngineSwitches(options string) []string {
	parser := shellwords.NewParser()
	parser.ParseEnv = false
	parser.ParseBacktick = false
	parser.ParseComment = false
	args, err := parser.Parse(options)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(args))
	for _, arg := range args {
		arg = strings.TrimPrefix(arg, consts.EngineSwitchPrefix)
		if arg != "" {
			out = append(out, arg)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
