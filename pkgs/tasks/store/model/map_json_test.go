package model

import (
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"testing"
)

func TestNullableStructJSON_nilIsSQLNull(t *testing.T) {
	got, err := NullableStructJSON[*taskcoredomain.PendingRetry](nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got %q want nil datatypes.JSON for SQL NULL", got)
	}
}

func TestNullableStructJSON_valueMarshals(t *testing.T) {
	v := &taskcoredomain.PendingRetry{Mode: taskcoredomain.RetryFresh, ParentCycleID: "cycle-1"}
	got, err := NullableStructJSON(v)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected marshaled json bytes")
	}
}
