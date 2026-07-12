package model

// CycleRow is a GORM association stub for task_cycles.id FK constraints.
type CycleRow struct {
	ID string `gorm:"primaryKey"`
}

// TableName pins the task_cycles table name for association metadata.
func (CycleRow) TableName() string { return "task_cycles" }
