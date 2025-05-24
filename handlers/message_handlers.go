package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"support_front_api/db"
	"support_front_api/logger"
	"support_front_api/models"
	"time"

	"github.com/gin-gonic/gin"
)

// AddMessage добавляет новое сообщение к тикету
func AddMessage(c *gin.Context) {
	ticketID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID тикета"})
		return
	}

	var request models.NewMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование тикета
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1)", ticketID).Scan(&exists)
	if err != nil {
		logger.LogError("Ошибка при проверке тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении сообщения"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Тикет не найден"})
		return
	}

	// Проверяем, что тикет не закрыт
	var status string
	var ticekt_user_id int
	err = db.DB.QueryRow("SELECT status, user_id FROM tickets WHERE id = $1", ticketID).Scan(&status, &ticekt_user_id)
	if err != nil {
		logger.LogError("Ошибка при получении статуса тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении сообщения"})
		return
	}

	if status == "закрыт" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя добавить сообщение в закрытый тикет"})
		return
	}

	// Добавляем сообщение
	var messageID int
	err = db.DB.QueryRow(
		"INSERT INTO ticket_messages (ticket_id, sender_type, sender_id, message, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		ticketID, request.SenderType, request.SenderID, request.Message, time.Now(),
	).Scan(&messageID)

	if err != nil {
		logger.LogError("Ошибка при добавлении сообщения: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении сообщения"})
		return
	}

	// Отправляем уведомление на localhost/superconnect
	go func() {
		superconnectURL := fmt.Sprintf("http://localhost/superconnect?super_connect_token=super_secret_key_2024&sender_id=5259653323&message=По вашему тикету %d пришло новое сообщение&accepter_id=5259653319", ticketID)

		resp, err := http.Post(superconnectURL, "application/x-www-form-urlencoded", nil)
		if err != nil {
			logger.LogError("Ошибка при отправке уведомления: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.LogError("Ошибка при отправке уведомления, код ответа: %d", resp.StatusCode)
		}
	}()

	go func() {

		data := url.Values{}
		data.Set("super_connect_token", "super_secret_key_2024")
		data.Set("sender_id", "5259653323")
		data.Set("message", "В вашем тиките "+strconv.Itoa(ticketID)+" обновление"+request.Message)
		data.Set("accepter_id", strconv.Itoa(ticekt_user_id))

		resp, err := http.PostForm("http://localhost:8443/superconnect", data)
		if err != nil {
			logger.LogError("Ошибка при отправке уведомления: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.LogError("Ошибка при отправке уведомления, код ответа: %d", resp.StatusCode)
		}
	}()
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Сообщение добавлено успешно",
		"message_id": messageID,
	})
}

// GetTicketMessages возвращает сообщения тикета
func GetTicketMessages(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении сообщений"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Тикет не найден"})
		return
	}

	// Получаем сообщения
	rows, err := db.DB.Query(
		"SELECT id, ticket_id, sender_type, sender_id, message, created_at FROM ticket_messages WHERE ticket_id = $1 ORDER BY created_at",
		ticketID,
	)
	if err != nil {
		logger.LogError("Ошибка при получении сообщений: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении сообщений"})
		return
	}
	defer rows.Close()

	var messages []models.TicketMessage
	for rows.Next() {
		var message models.TicketMessage
		if err := rows.Scan(
			&message.ID,
			&message.TicketID,
			&message.SenderType,
			&message.SenderID,
			&message.Message,
			&message.CreatedAt,
		); err != nil {
			logger.LogError("Ошибка при сканировании сообщения: %v", err)
			continue
		}
		messages = append(messages, message)
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}
