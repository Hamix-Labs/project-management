package execute

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/git"
)

// Service runs the execute I/O pipeline against explicit dependencies.
type Service struct {
	store     contract.Store
	git       *git.Service
	reportDir string
}

// Deps bundles Service construction inputs from harness root.
type Deps struct {
	Store     contract.Store
	Git       *git.Service
	ReportDir string
}

// NewService constructs an execute Service.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NewService(deps Deps) *Service {
	return &Service{
		store:     deps.Store,
		git:       deps.Git,
		reportDir: deps.ReportDir,
	}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) SetReportDir(dir string) {
	s.reportDir = dir
	if s.git != nil {
		s.git.SetReportDir(dir)
	}
}
