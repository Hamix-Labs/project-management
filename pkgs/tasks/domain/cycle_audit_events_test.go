package domain

import "testing"

func TestCycleEventTypeStringValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		got, want EventType
	}{
		{EventCycleStarted, "cycle_started"},
		{EventCycleCompleted, "cycle_completed"},
		{EventCycleFailed, "cycle_failed"},
		{EventPhaseStarted, "phase_started"},
		{EventPhaseCompleted, "phase_completed"},
		{EventPhaseFailed, "phase_failed"},
		{EventPhaseSkipped, "phase_skipped"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.want), func(t *testing.T) {
			t.Parallel()
			if string(tc.got) != string(tc.want) {
				t.Fatalf("event type drift: got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestEventTypeAcceptsUserResponse_excludesCycleAndPhaseEvents(t *testing.T) {
	t.Parallel()

	for _, et := range []EventType{
		EventCycleStarted,
		EventCycleCompleted,
		EventCycleFailed,
		EventPhaseStarted,
		EventPhaseCompleted,
		EventPhaseFailed,
		EventPhaseSkipped,
	} {
		et := et
		t.Run(string(et), func(t *testing.T) {
			t.Parallel()
			if EventTypeAcceptsUserResponse(et) {
				t.Fatalf("execution-cycle audit mirrors are observational; %q must not accept user_response", et)
			}
		})
	}
}
