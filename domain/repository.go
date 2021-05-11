package domain

import "github.com/google/uuid"

type Tasks []*Task

type Repository interface {
	Find(uuid.UUID) (*Task, error)
	Save(*Task) error
	FindAll() (Tasks, error)
}
