package mongodb

import (
	"GoNews/pkg/storage"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store представляет собой реализацию хранилища данных в MongoDB.
type Store struct {
	client     *mongo.Client
	Collection *mongo.Collection
	counters   *mongo.Collection
}

type Counter struct {
	ID  string `bson:"_id"` // Название счетчика
	Seq int    `bson:"seq"` // Значение счетчика
}

func getNextSequence(counterCollection *mongo.Collection, sequenceName string) (int, error) {
	filter := bson.M{"_id": sequenceName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	var counter Counter

	// Здесь используем counterCollection для выполнения операции FindOneAndUpdate
	err := counterCollection.FindOneAndUpdate(context.Background(), filter, update).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении следующего последовательного номера: %w", err)
	}

	return counter.Seq, nil
}

func New(uri, dbName, collectionName string) (*Store, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	}

	// Пинг до сервера MongoDB для проверки подключения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ошибка пинга к MongoDB: %w", err)
	}

	fmt.Println("Подключение к MongoDB успешно установлено.")

	// Проверка и создание коллекции счетчиков
	counterCollection := client.Database(dbName).Collection("counters")

	// Проверяем, существует ли счетчик для постов, если нет - создаем
	if _, err = counterCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": "postID"},
		bson.M{"$setOnInsert": bson.M{"seq": 0}}, // Устанавливаем начальное значение для seq, если счетчик не существует
		options.Update().SetUpsert(true),
	); err != nil {
		return nil, fmt.Errorf("ошибка проверки или создания счетчика постов: %w", err)
	}

	// Создаем основную коллекцию
	collection := client.Database(dbName).Collection(collectionName)

	return &Store{client: client, Collection: collection, counters: counterCollection}, nil
}

func (s *Store) Close() {
	if s.client != nil {
		if err := s.client.Disconnect(context.Background()); err != nil {
			fmt.Println("Ошибка при отключении от MongoDB:", err)
		} else {
			fmt.Println("Подключение к MongoDB закрыто.")
		}
	}
}

// Posts возвращает все публикации из базы данных.
func (s *Store) Posts() ([]storage.Post, error) {
	var posts []storage.Post

	cursor, err := s.Collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var post struct {
			ID          int    `bson:"id"`
			Title       string `bson:"title"`
			Content     string `bson:"content"`
			AuthorID    int    `bson:"author_id"`
			AuthorName  string `bson:"author_name"`
			CreatedAt   int64  `bson:"created_at"`
			PublishedAt int64  `bson:"published_at"`
		}

		if err := cursor.Decode(&post); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %w", err)
		}

		posts = append(posts, storage.Post{
			ID:          post.ID,
			Title:       post.Title,
			Content:     post.Content,
			AuthorID:    post.AuthorID,
			AuthorName:  post.AuthorName,
			CreatedAt:   post.CreatedAt,
			PublishedAt: post.PublishedAt,
		})
	}

	return posts, nil
}

// AddPost добавляет новую публикацию в базу данных.
func (s *Store) AddPost(post storage.Post) error {

	nextID, err := getNextSequence(s.counters, "postID")
	if err != nil {
		return err
	}

	newPost := struct {
		ID          int    `bson:"id"`
		Title       string `bson:"title"`
		Content     string `bson:"content"`
		AuthorID    int    `bson:"author_id"`
		AuthorName  string `bson:"author_name"`
		CreatedAt   int64  `bson:"created_at"`
		PublishedAt int64  `bson:"published_at"`
	}{
		ID:          nextID,
		Title:       post.Title,
		Content:     post.Content,
		AuthorID:    post.AuthorID,
		AuthorName:  post.AuthorName,
		CreatedAt:   post.CreatedAt,
		PublishedAt: post.PublishedAt,
	}

	_, err = s.Collection.InsertOne(context.Background(), newPost)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении поста: %w", err)
	}

	return nil
}

// UpdatePost обновляет существующую публикацию в базе данных.
func (s *Store) UpdatePost(post storage.Post) error {
	filter := bson.M{"id": post.ID} // Использовать ID как int
	update := bson.M{
		"$set": bson.M{
			"title":        post.Title,
			"content":      post.Content,
			"author_id":    post.AuthorID,
			"author_name":  post.AuthorName,
			"created_at":   post.CreatedAt,
			"published_at": post.PublishedAt,
		},
	}

	_, err := s.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении поста: %w", err)
	}

	return nil
}

// DeletePost удаляет публикацию из базы данных.
func (s *Store) DeletePost(post storage.Post) error {
	_, err := s.Collection.DeleteOne(context.Background(), bson.M{"id": post.ID})
	if err != nil {
		return fmt.Errorf("ошибка при удалении поста: %w", err)
	}

	return nil
}
