package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"support_front_api/db"
	"support_front_api/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadTicketPhoto загружает фотографию к тикету
func UploadTicketPhoto(c *gin.Context) {
	// Получаем ID тикета
	ticketID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID тикета"})
		return
	}

	// Проверяем существование тикета
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1)", ticketID).Scan(&exists)
	if err != nil {
		logger.LogError("Ошибка при проверке тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при загрузке фотографии"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Тикет не найден"})
		return
	}

	// Получаем данные отправителя
	senderType := c.PostForm("sender_type")
	senderID, err := strconv.ParseInt(c.PostForm("sender_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID отправителя"})
		return
	}

	// Получаем сообщение (опционально)
	messageIDStr := c.PostForm("message_id")
	var messageID sql.NullInt32

	if messageIDStr != "" {
		msgID, err := strconv.Atoi(messageIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID сообщения"})
			return
		}
		messageID.Int32 = int32(msgID)
		messageID.Valid = true
	}

	// Получаем файл
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось получить файл"})
		return
	}
	defer file.Close()

	// Создаем директорию для хранения файлов, если её нет
	uploadsDir := "../uploads/" + strconv.Itoa(ticketID)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.LogError("Ошибка при создании директории для загрузок: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при загрузке фотографии"})
		return
	}

	// Генерируем уникальное имя файла
	fileID := uuid.New().String()
	filename := fileID + filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadsDir, filename)

	// Сохраняем файл
	out, err := os.Create(filePath)
	if err != nil {
		logger.LogError("Ошибка при создании файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при загрузке фотографии"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		logger.LogError("Ошибка при копировании файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при загрузке фотографии"})
		return
	}

	// Сохраняем информацию о фотографии в базу данных
	var photoID int
	query := `
		INSERT INTO ticket_photos 
		(ticket_id, sender_type, sender_id, file_path, file_id, message_id, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`

	err = db.DB.QueryRow(
		query,
		ticketID,
		senderType,
		senderID,
		filePath,
		fileID,
		messageID,
		time.Now(),
	).Scan(&photoID)

	if err != nil {
		logger.LogError("Ошибка при сохранении информации о фотографии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при загрузке фотографии"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Фотография успешно загружена",
		"photo_id":  photoID,
		"file_id":   fileID,
		"file_path": filePath,
	})
}

// GetTicketPhoto получает фотографию тикета
func GetTicketPhoto(c *gin.Context) {
	// Получаем ID фотографии
	photoID, err := strconv.Atoi(c.Param("photo_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID фотографии"})
		return
	}

	// Получаем информацию о фотографии
	var filePath string
	err = db.DB.QueryRow(
		"SELECT file_path FROM ticket_photos WHERE id = $1",
		photoID,
	).Scan(&filePath)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Фотография не найдена"})
		} else {
			logger.LogError("Ошибка при получении информации о фотографии: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении фотографии"})
		}
		return
	}

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.LogError("Файл не найден: %v", filePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Файл фотографии не найден"})
		return
	}

	// Отправляем файл
	c.File(filePath)
}

// DeleteTicketPhoto удаляет фотографию тикета
func DeleteTicketPhoto(c *gin.Context) {
	// Получаем ID фотографии
	photoID, err := strconv.Atoi(c.Param("photo_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID фотографии"})
		return
	}

	// Получаем информацию о фотографии
	var filePath string
	err = db.DB.QueryRow(
		"SELECT file_path FROM ticket_photos WHERE id = $1",
		photoID,
	).Scan(&filePath)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Фотография не найдена"})
		} else {
			logger.LogError("Ошибка при получении информации о фотографии: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении фотографии"})
		}
		return
	}

	// Удаляем запись из базы данных
	_, err = db.DB.Exec("DELETE FROM ticket_photos WHERE id = $1", photoID)
	if err != nil {
		logger.LogError("Ошибка при удалении записи о фотографии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении фотографии"})
		return
	}

	// Удаляем файл
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			logger.LogWarning("Не удалось удалить файл фотографии: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Фотография успешно удалена",
		"photo_id": photoID,
	})
}
