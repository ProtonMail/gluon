package tests

import (
	"sync"

	"github.com/ProtonMail/gluon/reporter"
)

type testReporter struct {
	reports []report
	lock    sync.RWMutex
}

type report struct {
	val any
	ctx reporter.Context
}

func (r *testReporter) getReports() []report {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.reports
}

func (r *testReporter) ReportException(val any) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.reports = append(r.reports, report{val: val})

	return nil
}

func (r *testReporter) ReportMessage(val string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.reports = append(r.reports, report{val: val})

	return nil
}

func (r *testReporter) ReportMessageWithContext(val string, ctx reporter.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.reports = append(r.reports, report{val: val, ctx: ctx})

	return nil
}

func (r *testReporter) ReportExceptionWithContext(val any, ctx reporter.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.reports = append(r.reports, report{val: val, ctx: ctx})

	return nil
}
