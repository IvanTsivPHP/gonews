package memdb

import "GoNews/pkg/storage"

// Хранилище данных.
type Store struct{}

// Конструктор объекта хранилища.
func New() *Store {
	return new(Store)
}

func (s *Store) Posts() ([]storage.Post, error) {
	return posts, nil
}

func (s *Store) AddPost(storage.Post) error {
	return nil
}
func (s *Store) UpdatePost(storage.Post) error {
	return nil
}
func (s *Store) DeletePost(storage.Post) error {
	return nil
}

var posts = []storage.Post{
	{
		ID:      1,
		Title:   "Effective Go",
		Content: "wololo",
	},
	{
		ID:      2,
		Title:   "The Go Memory Model",
		Content: "The Go memory model specifies the conditions under which reads of a variable in one goroutine can be guaranteed to observe values produced by writes to the same variable in a different goroutine.",
	},
}
