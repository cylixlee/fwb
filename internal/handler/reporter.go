package handler

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// reporter implements domain.EventReporter for the CLI.
type reporter struct{}

func (reporter) Event(code string, args ...any) {
	fmt.Printf(consts.FormatLogEvent, Translate(code, args...))
}

func (reporter) StartProgress(code string, args ...any) domain.OperationProgress {
	label := consts.ProgressDescriptionPrefix + Translate(consts.ProgressKeyPrefix+code, args...)
	bar := progressbar.NewOptions64(
		-1,
		progressbar.OptionSetDescription(label),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetWidth(consts.ProgressBarWidth),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(consts.ProgressThrottleInterval),
		progressbar.OptionSpinnerType(consts.ProgressSpinnerType),
		progressbar.OptionEnableColorCodes(false),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprintln(os.Stdout)
		}),
	)
	return &downloadProgress{bar: bar}
}

type downloadProgress struct {
	bar   *progressbar.ProgressBar
	total int64
}

func (p *downloadProgress) Update(done, total int64) {
	if total > 0 && total != p.total {
		p.total = total
		p.bar.ChangeMax64(total)
	}
	_ = p.bar.Set64(done)
}

func (p *downloadProgress) Done() {
	_ = p.bar.Finish()
}
