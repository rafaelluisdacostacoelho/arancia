package todo

// ToDo represents a todo item
type ToDo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// CreateToDoRequest represents the request body for creating a todo
type CreateToDoRequest struct {
	Title string `json:"title"`
}

// UpdateToDoRequest represents the request body for updating a todo
type UpdateToDoRequest struct {
	Title     *string `json:"title,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
}
