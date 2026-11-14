package handler

import (
	"fmt"
	"io"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"golang.org/x/time/rate"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// reporter implements domain.EventReporter for the CLI.
type reporter struct {
	out         io.Writer
	interactive bool
}

// newReporter builds a reporter for out. Only a terminal gets the in-place
// progress bar: the bar animates with carriage returns, which every other
// consumer renders as one line per frame.
func newReporter(out io.Writer, interactive bool) *reporter {
	return &reporter{out: out, interactive: interactive}
}

// NewReporter returns the event reporter bound to the process stdout.
func NewReporter() domain.EventReporter {
	return newReporter(os.Stdout, term.IsTerminal(int(os.Stdout.Fd())))
}

func (r *reporter) Event(code string, args ...any) {
	fmt.Fprintf(r.out, consts.FormatLogEvent, Translate(code, args...))
}

func (r *reporter) StartProgress(code string, args ...any) domain.OperationProgress {
	text := Translate(consts.ProgressKeyPrefix+code, args...)
	if !r.interactive {
		return &lineProgress{
			out:     r.out,
			label:   text,
			limiter: rate.NewLimiter(rate.Every(consts.ProgressLogInterval), 1),
		}
	}
	bar := progressbar.NewOptions64(
		-1,
		progressbar.OptionSetDescription(consts.ProgressDescriptionPrefix+text),
		progressbar.OptionSetWriter(r.out),
		progressbar.OptionSetWidth(consts.ProgressBarWidth),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(consts.ProgressThrottleInterval),
		progressbar.OptionSpinnerType(consts.ProgressSpinnerType),
		progressbar.OptionEnableColorCodes(false),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(r.out)
		}),
	)
	return &barProgress{bar: bar}
}

// barProgress renders an in-place progress bar for terminals.
type barProgress struct {
	bar   *progressbar.ProgressBar
	total int64
}

func (p *barProgress) Update(done, total int64) {
	if total > 0 && total != p.total {
		p.total = total
		p.bar.ChangeMax64(total)
	}
	p.bar.Set64(done)
}

func (p *barProgress) Done() {
	p.bar.Finish()
}

// Abort keeps the partial line instead of claiming completion.
func (p *barProgress) Abort() {
	p.bar.Exit()
}

// lineProgress renders one line per interval for non-terminal writers.
type lineProgress struct {
	out     io.Writer
	label   string
	done    int64
	total   int64
	printed int64
	limiter *rate.Limiter
}

func (p *lineProgress) Update(done, total int64) {
	p.done = done
	if total > 0 {
		p.total = total
	}
	if p.total <= 0 || !p.limiter.Allow() {
		return
	}
	p.print()
}

func (p *lineProgress) Done() {
	// Skip the final line when it would repeat the last one.
	if p.total <= 0 || p.done == p.printed {
		return
	}
	p.print()
}

// Abort prints nothing: the caller reports the failure.
func (p *lineProgress) Abort() {}

func (p *lineProgress) print() {
	p.printed = p.done
	fmt.Fprintf(p.out, consts.FormatLogEvent, Translate(consts.KeyProgressLine,
		p.label,
		p.done*consts.ProgressPercentScale/p.total,
		humanize.Bytes(uint64(p.done)),
		humanize.Bytes(uint64(p.total)),
	))
}
