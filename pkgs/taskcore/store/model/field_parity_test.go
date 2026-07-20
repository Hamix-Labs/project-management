package model

import (
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/storekernel/parity"
)

func TestFieldParity(t *testing.T) {
	t.Parallel()
	for _, pair := range ParityPairs {
		pair := pair
		t.Run(pair.Name, func(t *testing.T) {
			t.Parallel()
			if err := parity.AssertFieldParity(pair); err != nil {
				t.Fatal(err)
			}
		})
	}
}
