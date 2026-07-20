package scheduling

import (
	"time"

	taskcorecontract "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
)

// ShouldNotifyReadyNow returns true when a freshly-ready task may enter the
// in-memory queue immediately. Only pickup_not_before is checked — see ADR-0023 I7.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ShouldNotifyReadyNow(pickupNotBefore *time.Time, now time.Time) bool {
	return taskcorecontract.ShouldNotifyReadyNow(pickupNotBefore, now)
}
