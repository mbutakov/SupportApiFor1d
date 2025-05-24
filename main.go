package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"support_front_api/config"
	"support_front_api/db"
	"support_front_api/handlers"
	"support_front_api/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Парсим флаги командной строки
	configPath := flag.String("config", "config/.config", "путь к файлу конфигурации")
	flag.Parse()

	// Загружаем или создаем конфигурацию
	var cfg *config.Config
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		// Создаем директорию для конфигурационного файла
		configDir := filepath.Dir(*configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("Ошибка при создании директории для конфигурации: %v", err)
		}

		// Создаем конфигурацию по умолчанию
		cfg = config.DefaultConfig()
		if err := config.SaveConfig(cfg, *configPath); err != nil {
			log.Fatalf("Ошибка при сохранении конфигурации: %v", err)
		}
		log.Printf("Создана конфигурация по умолчанию в %s", *configPath)
	} else {
		// Загружаем существующую конфигурацию
		var err error
		cfg, err = config.LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
		}
		log.Printf("Загружена конфигурация из %s", *configPath)
	}

	// Инициализируем логирование
	if err := logger.InitLogger(cfg.LogFilePath); err != nil {
		log.Fatalf("Ошибка при инициализации логирования: %v", err)
	}

	// Инициализируем соединение с базой данных
	if err := db.InitDB(cfg); err != nil {
		logger.LogError("Ошибка при инициализации базы данных: %v", err)
		log.Fatalf("Ошибка при инициализации базы данных: %v", err)
	}
	defer db.CloseDB()

	// Создаем директорию для загрузок
	uploadsDir := "./uploads/"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.LogError("Ошибка при создании директории для загрузок: %v", err)
		log.Fatalf("Ошибка при создании директории для загрузок: %v", err)
	}

	// Инициализация роутера Gin
	router := gin.Default()

	// Увеличиваем максимальный размер загружаемых файлов
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Настройка CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.AllowOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Статические файлы
	router.Static("/uploads", "../uploads")
	router.Static("/api/uploads", "../uploads")
	// Базовый маршрут для проверки работы API
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API для тикетов поддержки работает",
		})
	})

	// Группа маршрутов для тикетов
	ticketsGroup := router.Group("/api/tickets")
	{
		ticketsGroup.GET("/", handlers.GetAllTickets)
		ticketsGroup.GET("/:id", handlers.GetTicketById)
		ticketsGroup.POST("/", handlers.CreateTicket)
		ticketsGroup.PUT("/:id", handlers.UpdateTicket)
		ticketsGroup.DELETE("/:id", handlers.DeleteTicket)

		// Маршруты для сообщений в тикетах
		ticketsGroup.POST("/:id/messages", handlers.AddMessage)
		ticketsGroup.GET("/:id/messages", handlers.GetTicketMessages)

		// Маршруты для фотографий в тикетах
		ticketsGroup.POST("/:id/photos", handlers.UploadTicketPhoto)
		ticketsGroup.GET("/photos/:photo_id", handlers.GetTicketPhoto)
		ticketsGroup.DELETE("/photos/:photo_id", handlers.DeleteTicketPhoto)
	}

	// Группа маршрутов для пользователей
	usersGroup := router.Group("/api/users")
	{
		usersGroup.GET("/", handlers.GetUsers)
		usersGroup.GET("/:id", handlers.GetUserById)
		usersGroup.POST("/", handlers.CreateUser)
		usersGroup.PUT("/:id", handlers.UpdateUser)
	}

	// Запуск сервера
	logger.LogInfo("Запуск сервера на порту %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.LogError("Ошибка запуска сервера: %v", err)
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
