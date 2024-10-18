package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"GoNews/pkg/storage"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Store представляет собой реализацию хранилища данных в PostgreSQL.
var DBPool *pgxpool.Pool

type Store struct {
	db *pgxpool.Pool
}

func InitDB(dsn string) error {
	// Устанавливаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Парсим строку подключения
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("не удалось разобрать строку подключения: %w", err)
	}

	// Подключаемся к базе данных
	DBPool, err = pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	fmt.Println("Подключение к базе данных успешно установлено.")
	return nil
}

// CloseDB закрывает глобальное соединение для миграций или других целей.
func CloseDB() {
	if DBPool != nil {
		DBPool.Close()
		fmt.Println("Подключение к базе данных закрыто.")
	}
}

func New(dsn string) (*Store, error) {
	err := InitDB(dsn)
	if err != nil {
		return nil, err
	}

	// Возвращаем объект Store с использованием глобального пула DBPool
	return &Store{db: DBPool}, nil
}

// Close закрывает пул соединений для основного хранилища.
func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
		fmt.Println("Подключение к базе данных закрыто для основного хранилища.")
	}
}

// Posts возвращает все публикации из базы данных, включая информацию об авторах.
func (s *Store) Posts() ([]storage.Post, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT p.id, p.title, p.content, p.author_id, a.name, p.created_at 
		FROM posts p
		JOIN authors a ON p.author_id = a.id`)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var posts []storage.Post
	for rows.Next() {
		var post storage.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.AuthorName, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// AddPost добавляет новую публикацию в базу данных.
func (s *Store) AddPost(post storage.Post) error {
	// Проверка на существование автора
	var authorExists bool
	err := s.db.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM authors WHERE id = $1)`, post.AuthorID).Scan(&authorExists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования автора: %w", err)
	}

	if !authorExists {
		return fmt.Errorf("автор с ID %d не существует", post.AuthorID)
	}

	_, err = s.db.Exec(context.Background(),
		`INSERT INTO posts (title, content, author_id, created_at) VALUES ($1, $2, $3, $4)`,
		post.Title, post.Content, post.AuthorID, time.Now().Unix())

	if err != nil {
		return fmt.Errorf("ошибка при добавлении поста: %w", err)
	}

	return nil
}

// UpdatePost обновляет существующую публикацию в базе данных.
func (s *Store) UpdatePost(post storage.Post) error {
	// Сначала создаем переменную для создания запроса
	query := `UPDATE posts SET`
	var args []interface{}
	var setClauses []string

	if post.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf(" title = $%d", len(args)+1))
		args = append(args, post.Title)
	}

	if post.Content != "" {
		setClauses = append(setClauses, fmt.Sprintf(" content = $%d", len(args)+1))
		args = append(args, post.Content)
	}

	if post.AuthorID != 0 {
		setClauses = append(setClauses, fmt.Sprintf(" author_id = $%d", len(args)+1))
		args = append(args, post.AuthorID)
	}

	if post.CreatedAt != 0 {
		setClauses = append(setClauses, fmt.Sprintf(" created_at = $%d", len(args)+1))
		args = append(args, post.CreatedAt)
	}

	// Если не указаны поля для обновления, возвращаем ошибку
	if len(setClauses) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}

	// Добавляем условие WHERE к запросу
	query += fmt.Sprintf("%s WHERE id = $%d",
		strings.Join(setClauses, ","),
		len(args)+1)

	// Добавляем ID поста в аргументы
	args = append(args, post.ID)

	// Выполняем запрос
	_, err := s.db.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении поста: %w", err)
	}

	return nil
}

// DeletePost удаляет публикацию из базы данных.
func (s *Store) DeletePost(post storage.Post) error {
	_, err := s.db.Exec(context.Background(), `DELETE FROM posts WHERE id = $1`, post.ID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении поста: %w", err)
	}

	return nil
}
