package domain

// ProjectStatus is the lifecycle state of a long-lived project.
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)
