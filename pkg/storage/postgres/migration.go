package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

func Migrate(dsn string) error {

	// Инициализируем подключение к базе данных
	if err := InitDB(dsn); err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}
	defer CloseDB() // Закрываем подключение в конце

	// Открываем файл с миграциями
	file, err := os.Open("schema.sql")
	if err != nil {
		return fmt.Errorf("не удалось открыть файл schema.sql: %w", err)
	}
	defer file.Close()

	// Читаем содержимое файла с миграциями
	sqlBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла schema.sql: %w", err)
	}
	sql := string(sqlBytes)

	// Настраиваем контекст для выполнения запроса
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполняем SQL-скрипт миграции
	_, err = DBPool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %w", err)
	}

	// Сообщаем об успешной миграции
	fmt.Println("Миграция выполнена успешно!")
	return nil
}
