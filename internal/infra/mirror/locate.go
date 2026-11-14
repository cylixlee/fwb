package mirror

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/pkg/fsx"
)

func (m *projectMirror) Locate(mode string) (string, string, error) {
	pattern := filepath.Join(m.root, consts.BuildDirectory, consts.WindowsDirectory, consts.GlobWildcard, consts.RunnerDirectory, mode)
	matches, _ := filepath.Glob(pattern)
	sort.Strings(matches)
	for _, match := range slices.Backward(matches) {
		if exe := fsx.FindExecutable(match, m.binaryName(), consts.ExeSuffix); exe != "" {
			return match, exe, nil
		}
	}
	return "", "", domain.NewError(consts.ErrorCodeProjectNoBuild, pattern)
}

func (m *projectMirror) binaryName() string {
	data, err := os.ReadFile(filepath.Join(m.root, consts.WindowsDirectory, consts.CMakeListsFileName))
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), consts.LineFeed) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(line), consts.BinaryNameDirective) {
			continue
		}
		start := strings.Index(line, consts.DoubleQuote)
		end := strings.LastIndex(line, consts.DoubleQuote)
		if start >= 0 && end > start {
			return line[start+1 : end]
		}
	}
	return ""
}
