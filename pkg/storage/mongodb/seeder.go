package mongodb

import (
	"GoNews/pkg/storage"
	"fmt"
	"time"
)

// Seeder заполняет коллекцию начальными данными и возвращает список ошибок.
func SeedPosts(store Store) []error {
	var errors []error

	// Данные для постов
	posts := []storage.Post{
		{
			Title:       "Первый пост",
			Content:     "Содержимое первого поста.",
			AuthorID:    1,
			AuthorName:  "Автор 1",
			CreatedAt:   time.Now().Unix(),
			PublishedAt: time.Now().Unix(),
		},
		{
			Title:       "Второй пост",
			Content:     "Содержимое второго поста.",
			AuthorID:    2,
			AuthorName:  "Автор 2",
			CreatedAt:   time.Now().Unix(),
			PublishedAt: time.Now().Unix(),
		},
		{
			Title:       "Третий пост",
			Content:     "Содержимое третьего поста.",
			AuthorID:    3,
			AuthorName:  "Автор 3",
			CreatedAt:   time.Now().Unix(),
			PublishedAt: time.Now().Unix(),
		},
	}

	for _, post := range posts {
		err := store.AddPost(post)
		if err != nil {
			// Добавляем ошибку в список ошибок
			errors = append(errors, fmt.Errorf("ошибка при добавлении поста '%s': %v", post.Title, err))
		} else {
			fmt.Printf("Пост '%s' успешно добавлен.\n", post.Title)
		}
	}

	return errors
}
