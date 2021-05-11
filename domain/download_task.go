package domain

import "github.com/google/uuid"

type TaskStatus string

const (
	StatusNew     TaskStatus = "NEW"
	StatusPending TaskStatus = "PENDING"
	StatusDone    TaskStatus = "DONE"
	StatusFailed  TaskStatus = "FAILED"
)

type Task struct {
	ID          uuid.UUID
	OriginURL   string
	DownloadURL string
	Status      TaskStatus
}
