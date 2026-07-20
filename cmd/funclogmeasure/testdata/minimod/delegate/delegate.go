package delegate

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

func ViaRunObserved(ctx context.Context) error {
	return calltrace.RunObserved(ctx, "test.op", nil, func(ctx context.Context) ([]any, error) {
		return nil, nil
	})
}

func innerLogged(ctx context.Context) error {
	return calltrace.RunObserved(ctx, "inner.op", nil, func(ctx context.Context) ([]any, error) {
		return nil, nil
	})
}

func ThinWrapper(ctx context.Context) error {
	return innerLogged(ctx)
}
