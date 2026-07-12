package model

// TaskRow is a GORM association stub for tasks.id FK constraints.
type TaskRow struct {
	ID string `gorm:"primaryKey"`
}

// TableName pins the tasks table name for association metadata.
func (TaskRow) TableName() string { return "tasks" }
