package main

import (
	"GoNews/pkg/api"
	"GoNews/pkg/storage"
	"GoNews/pkg/storage/memdb"
	"GoNews/pkg/storage/mongodb"
	"GoNews/pkg/storage/postgres"
	"flag"
	"fmt"
	"log"
	"net/http"

	"GoNews/config"

	"github.com/spf13/viper"
)

// Сервер GoNews.
type server struct {
	db  storage.Interface
	api *api.API
}

func main() {

	// Загрузка конфигурации с помощью Viper
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	var srv server

	migrate := flag.Bool("migrate", false, "Run database migrations")
	seed := flag.Bool("seed", false, "Seed the database with initial data") // Флаг для сидирования
	dbType := flag.String("db", "memdb", "Specify the database type: postgres, memdb, mongodb")
	flag.Parse()

	switch *dbType {
	case "postgres":
		if *migrate {
			fmt.Println("Запуск миграции для PostgreSQL...")
			err := postgres.Migrate(cfg.GetPostgresDSN())
			if err != nil {
				log.Fatalf("Ошибка при выполнении миграции: %v", err)
			}
			fmt.Println("Миграция завершена.")
		}
		srv.db, err = postgres.New(cfg.GetPostgresDSN())
		if err != nil {
			log.Fatalf("Ошибка при инициализации базы данных PostgreSQL: %v", err)
		}

	case "memdb":
		srv.db = memdb.New()

	case "mongodb":
		mongoDB, err := mongodb.New(cfg.Database.MongoDB.URI, cfg.Database.MongoDB.Name, "posts")
		if err != nil {
			log.Fatalf("Ошибка при инициализации MongoDB: %v", err)
		}
		defer mongoDB.Close()

		if *seed {
			// Сначала проверим, нужно ли сидировать
			fmt.Println("Запуск сидирования для MongoDB...")

			err := mongodb.SeedPosts(*mongoDB)
			if err != nil {
				log.Fatalf("Ошибка при сидировании базы данных: %v", err)
			}
			fmt.Println("Сидирование завершено.")
		}
		srv.db = mongoDB

	default:
		log.Fatalf("Неизвестный тип базы данных: %s", cfg.Database.Type)
	}

	// Создаём объект API и регистрируем обработчики.
	srv.api = api.New(srv.db)

	// Запускаем веб-сервер на порту 8080 на всех интерфейсах.
	// Предаём серверу маршрутизатор запросов,
	// поэтому сервер будет все запросы отправлять на маршрутизатор.
	// Маршрутизатор будет выбирать нужный обработчик.

	port := viper.GetInt("server.port")
	addr := fmt.Sprintf(":%d", port)

	http.ListenAndServe(addr, srv.api.Router())
}
