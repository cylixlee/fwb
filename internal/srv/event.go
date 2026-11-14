package srv

import (
	"github.com/cylixlee/fwb/internal/domain"
)

// emit reports an application event through rep.
func emit(rep domain.EventReporter, code string, args ...any) {
	if rep != nil {
		rep.Event(code, args...)
	}
}

// meter starts a progress meter; a nil reporter yields a no-op meter.
func meter(rep domain.EventReporter, code string, args ...any) domain.OperationProgress {
	if rep == nil {
		return nopProgress{}
	}
	return rep.StartProgress(code, args...)
}

type nopProgress struct{}

func (nopProgress) Update(int64, int64) {}
func (nopProgress) Done()               {}
func (nopProgress) Abort()              {}
