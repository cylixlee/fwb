package handler

import (
	"fmt"
	"time"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/dustin/go-humanize"
)

// reporter implements domain.EventReporter for the CLI.
type reporter struct{}

func (reporter) Event(code string, args ...any) {
	fmt.Printf(consts.FormatLogEvent, Translate(code, args...))
}

func (reporter) StartProgress(code string, args ...any) domain.OperationProgress {
	return &downloadProgress{
		label:   Translate(consts.ProgressKeyPrefix+code, args...),
		started: time.Now(),
		lastPct: -1,
	}
}

type downloadProgress struct {
	label   string
	started time.Time
	last    time.Time
	lastPct int
}

func (p *downloadProgress) Update(done, total int64) {
	now := time.Now()
	pct := int64(-1)
	if total > 0 {
		pct = done * 100 / total
	}
	if pct >= 0 {
		if pct == int64(p.lastPct) {
			return
		}
		if p.lastPct >= 0 && now.Sub(p.last) < consts.ProgressThrottleInterval {
			return
		}
		p.lastPct = int(pct)
	} else if now.Sub(p.last) < consts.ProgressThrottleInterval {
		return
	}
	p.last = now
	elapsed := now.Sub(p.started).Seconds()
	speed := int64(0)
	if elapsed > 0 {
		speed = int64(float64(done) / elapsed)
	}
	if total > 0 {
		fmt.Printf(consts.FormatLogEvent, Translate(consts.KeyProgressDownloading, p.label, pct, humanize.IBytes(uint64(done)), humanize.IBytes(uint64(total)), humanize.IBytes(uint64(speed))))
	} else {
		fmt.Printf(consts.FormatLogEvent, Translate(consts.KeyProgressDownloadingUnknown, p.label, humanize.IBytes(uint64(done)), humanize.IBytes(uint64(speed))))
	}
}

func (p *downloadProgress) Done() {
	fmt.Printf(consts.FormatLogEvent, Translate(consts.KeyProgressDone, p.label))
}
