package domain

import (
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
)

// ParseEngineSwitches splits the flutter tool's `${engineOptions}` value into
// individual engine switches. The value arrives as one argv element because the
// custom-device hook is exec'd without a shell. Quoted values keep spaces.
func ParseEngineSwitches(options string) (out []string) {
	var cur strings.Builder
	quote := rune(0)
	defer func() {
		if cur.Len() > 0 {
			out = append(out, strings.TrimPrefix(cur.String(), consts.EngineSwitchPrefix))
		}
	}()
	for _, r := range options {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if cur.Len() > 0 {
				out = append(out, strings.TrimPrefix(cur.String(), consts.EngineSwitchPrefix))
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	return out
}
