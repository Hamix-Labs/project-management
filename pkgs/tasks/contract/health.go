package contract

import "context"

// HealthStore covers readiness and coarse inventory checks for /health routes.
type HealthStore interface {
	Ready(ctx context.Context) error
	CountGitRepositories(ctx context.Context) (int64, error)
}
