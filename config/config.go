package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config структура для конфигурации
type Config struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`
	Database struct {
		Type     string         `mapstructure:"type"` // Тип базы данных
		Postgres PostgresConfig `mapstructure:"postgres"`
		MongoDB  MongoDBConfig  `mapstructure:"mongodb"`
		MemDB    struct{}       `mapstructure:"memdb"` // Пустая структура для memdb
	} `mapstructure:"database"`
}

// PostgresConfig структура для конфигурации PostgreSQL
type PostgresConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	SSLMode  string `mapstructure:"sslmode"`
}

// MongoDBConfig структура для конфигурации MongoDB
type MongoDBConfig struct {
	URI  string `mapstructure:"uri"`
	Name string `mapstructure:"dbname"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig() (Config, error) {
	var cfg Config
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("не удалось прочитать конфигурацию: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("не удалось распаковать конфигурацию: %w", err)
	}

	return cfg, nil
}

func (cfg *Config) GetPostgresDSN() string {
	// Извлекаем данные из PostgresConfig
	postgres := cfg.Database.Postgres

	// Формируем строку DSN
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		postgres.User,
		postgres.Password,
		postgres.Host,
		postgres.Port,
		postgres.Name,
		postgres.SSLMode,
	)

	return dsn
}
