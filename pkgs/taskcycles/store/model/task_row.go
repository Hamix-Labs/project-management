package model

// TaskRow is a GORM association stub for tasks.id FK constraints.
// The full Task model lives in pkgs/taskcore/store/model.
type TaskRow struct {
	ID string `gorm:"primaryKey"`
}

// TableName pins the tasks table name for association metadata.
func (TaskRow) TableName() string { return "tasks" }
