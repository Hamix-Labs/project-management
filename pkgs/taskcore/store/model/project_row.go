package model

// ProjectRow is a GORM association stub for projects.id FK constraints.
type ProjectRow struct {
	ID string `gorm:"primaryKey"`
}

// TableName pins the projects table name for association metadata.
func (ProjectRow) TableName() string { return "projects" }
