package db

import (
	"database/sql"
	"fmt"
	"support_front_api/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB инициализирует подключение к базе данных
func InitDB(cfg *config.Config) error {
	var err error
	DB, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка проверки соединения с базой данных: %v", err)
	}

	return nil
}

// CloseDB закрывает соединение с базой данных
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
