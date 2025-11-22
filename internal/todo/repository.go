package todo

import "errors"

// Repository defines the interface for todo storage operations
type Repository interface {
	GetAll() ([]*ToDo, error)
	GetByID(id string) (*ToDo, error)
	Create(todo *ToDo) error
	Update(id string, todo *ToDo) error
	Delete(id string) error
	Close() error
}

// Common errors
var (
	ErrNotFound = errors.New("todo not found")
	ErrInvalid  = errors.New("invalid todo data")
)

