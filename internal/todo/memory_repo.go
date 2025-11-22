package todo

import (
	"sync"
)

// MemoryRepository is an in-memory implementation of Repository
type MemoryRepository struct {
	mu    sync.RWMutex
	todos map[string]*ToDo
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		todos: make(map[string]*ToDo),
	}
}

// GetAll returns all todos
func (r *MemoryRepository) GetAll() ([]*ToDo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ToDo, 0, len(r.todos))
	for _, todo := range r.todos {
		result = append(result, todo)
	}
	return result, nil
}

// GetByID returns a todo by ID
func (r *MemoryRepository) GetByID(id string) (*ToDo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todo, exists := r.todos[id]
	if !exists {
		return nil, ErrNotFound
	}
	return todo, nil
}

// Create adds a new todo
func (r *MemoryRepository) Create(todo *ToDo) error {
	if todo.ID == "" || todo.Title == "" {
		return ErrInvalid
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[todo.ID]; exists {
		return ErrInvalid
	}

	r.todos[todo.ID] = todo
	return nil
}

// Update updates an existing todo
func (r *MemoryRepository) Update(id string, todo *ToDo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[id]; !exists {
		return ErrNotFound
	}

	r.todos[id] = todo
	return nil
}

// Delete removes a todo
func (r *MemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[id]; !exists {
		return ErrNotFound
	}

	delete(r.todos, id)
	return nil
}

// Close is a no-op for in-memory repository
func (r *MemoryRepository) Close() error {
	return nil
}

