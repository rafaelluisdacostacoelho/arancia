package todo

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

const todosBucket = "todos"

// BoltDBRepository is a BoltDB implementation of Repository
type BoltDBRepository struct {
	db *bbolt.DB
}

// NewBoltDBRepository creates a new BoltDB repository
func NewBoltDBRepository(path string) (*BoltDBRepository, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt db: %w", err)
	}

	// Create bucket if it doesn't exist
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(todosBucket))
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return &BoltDBRepository{db: db}, nil
}

// GetAll returns all todos
func (r *BoltDBRepository) GetAll() ([]*ToDo, error) {
	var todos []*ToDo

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(todosBucket))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var todo ToDo
			if err := json.Unmarshal(v, &todo); err != nil {
				return err
			}
			todos = append(todos, &todo)
			return nil
		})
	})

	return todos, err
}

// GetByID returns a todo by ID
func (r *BoltDBRepository) GetByID(id string) (*ToDo, error) {
	var todo *ToDo

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(todosBucket))
		if bucket == nil {
			return ErrNotFound
		}

		data := bucket.Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}

		todo = &ToDo{}
		return json.Unmarshal(data, todo)
	})

	if err != nil {
		return nil, err
	}
	return todo, nil
}

// Create adds a new todo
func (r *BoltDBRepository) Create(todo *ToDo) error {
	if todo.ID == "" || todo.Title == "" {
		return ErrInvalid
	}

	return r.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(todosBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Check if already exists
		if bucket.Get([]byte(todo.ID)) != nil {
			return ErrInvalid
		}

		data, err := json.Marshal(todo)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(todo.ID), data)
	})
}

// Update updates an existing todo
func (r *BoltDBRepository) Update(id string, todo *ToDo) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(todosBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		// Check if exists
		if bucket.Get([]byte(id)) == nil {
			return ErrNotFound
		}

		data, err := json.Marshal(todo)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(id), data)
	})
}

// Delete removes a todo
func (r *BoltDBRepository) Delete(id string) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(todosBucket))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		if bucket.Get([]byte(id)) == nil {
			return ErrNotFound
		}

		return bucket.Delete([]byte(id))
	})
}

// Close closes the database connection
func (r *BoltDBRepository) Close() error {
	return r.db.Close()
}

