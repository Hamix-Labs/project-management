package handler

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
)

var _ contract.AgentWorkerControl = (*fakeAgentControl)(nil)

type fakeAgentControl struct {
	cancelResult  atomic.Bool
	cancelCalls   atomic.Int32
	reloadCalls   atomic.Int32
	reloadErr     atomic.Pointer[error]
	probeCalls    atomic.Int32
	probeVersion  atomic.Pointer[string]
	probeResolved atomic.Pointer[string]
	probeErr      atomic.Pointer[error]
	lastRunner    atomic.Pointer[string]
	lastBinary    atomic.Pointer[string]
}

func (f *fakeAgentControl) CancelCurrentRun() bool {
	f.cancelCalls.Add(1)
	return f.cancelResult.Load()
}

func (f *fakeAgentControl) Reload(_ context.Context) error {
	f.reloadCalls.Add(1)
	if e := f.reloadErr.Load(); e != nil {
		return *e
	}
	return nil
}

func (f *fakeAgentControl) ProbeRunner(_ context.Context, runnerID, binaryPath string, _ time.Duration) (string, string, error) {
	f.probeCalls.Add(1)
	r := runnerID
	b := binaryPath
	f.lastRunner.Store(&r)
	f.lastBinary.Store(&b)
	resolved := ""
	if rp := f.probeResolved.Load(); rp != nil {
		resolved = *rp
	}
	if e := f.probeErr.Load(); e != nil {
		return "", resolved, *e
	}
	if v := f.probeVersion.Load(); v != nil {
		return *v, resolved, nil
	}
	return "", resolved, nil
}
