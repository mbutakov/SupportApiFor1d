package handlers

import (
	"database/sql"
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

// GetAllTickets возвращает список всех тикетов
func GetAllTickets(c *gin.Context) {
	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Фильтрация по статусу
	status := c.Query("status")

	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = db.DB.Query(
			"SELECT id, user_id, title, description, status, category, created_at, closed_at FROM tickets WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			status, limit, offset,
		)
	} else {
		rows, err = db.DB.Query(
			"SELECT id, user_id, title, description, status, category, created_at, closed_at FROM tickets ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			limit, offset,
		)
	}

	if err != nil {
		logger.LogError("Ошибка при получении тикетов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении тикетов"})
		return
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var closedAt sql.NullTime

		if err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.Title,
			&ticket.Description,
			&ticket.Status,
			&ticket.Category,
			&ticket.CreatedAt,
			&closedAt,
		); err != nil {
			logger.LogError("Ошибка при сканировании строки тикета: %v", err)
			continue
		}

		if closedAt.Valid {
			closedAtTime := closedAt.Time
			ticket.ClosedAt = &closedAtTime
		}

		tickets = append(tickets, ticket)
	}

	// Получаем общее количество тикетов для пагинации
	var count int
	var countErr error

	if status != "" {
		countErr = db.DB.QueryRow("SELECT COUNT(*) FROM tickets WHERE status = $1", status).Scan(&count)
	} else {
		countErr = db.DB.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&count)
	}

	if countErr != nil {
		logger.LogError("Ошибка при подсчете тикетов: %v", countErr)
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets": tickets,
		"total":   count,
		"page":    page,
		"limit":   limit,
	})
}

// GetTicketById возвращает тикет по ID
func GetTicketById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID тикета"})
		return
	}

	var ticket models.Ticket
	var closedAt sql.NullTime

	err = db.DB.QueryRow(
		"SELECT id, user_id, title, description, status, category, created_at, closed_at FROM tickets WHERE id = $1",
		id,
	).Scan(
		&ticket.ID,
		&ticket.UserID,
		&ticket.Title,
		&ticket.Description,
		&ticket.Status,
		&ticket.Category,
		&ticket.CreatedAt,
		&closedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Тикет не найден"})
		} else {
			logger.LogError("Ошибка при получении тикета: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении тикета"})
		}
		return
	}

	if closedAt.Valid {
		closedAtTime := closedAt.Time
		ticket.ClosedAt = &closedAtTime
	}

	// Получаем сообщения тикета
	rows, err := db.DB.Query(
		"SELECT id, ticket_id, sender_type, sender_id, message, created_at FROM ticket_messages WHERE ticket_id = $1 ORDER BY created_at",
		id,
	)
	if err != nil {
		logger.LogError("Ошибка при получении сообщений тикета: %v", err)
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
			logger.LogError("Ошибка при сканировании строки сообщения: %v", err)
			continue
		}
		messages = append(messages, message)
	}

	// Получаем фотографии тикета
	photoRows, err := db.DB.Query(
		"SELECT id, ticket_id, sender_type, sender_id, file_path, file_id, message_id, created_at FROM ticket_photos WHERE ticket_id = $1",
		id,
	)
	if err != nil {
		logger.LogError("Ошибка при получении фотографий тикета: %v", err)
	}
	defer photoRows.Close()

	var photos []models.TicketPhoto
	for photoRows.Next() {
		var photo models.TicketPhoto
		var messageID sql.NullInt32

		if err := photoRows.Scan(
			&photo.ID,
			&photo.TicketID,
			&photo.SenderType,
			&photo.SenderID,
			&photo.FilePath,
			&photo.FileID,
			&messageID,
			&photo.CreatedAt,
		); err != nil {
			logger.LogError("Ошибка при сканировании строки фотографии: %v", err)
			continue
		}

		if messageID.Valid {
			msgID := int(messageID.Int32)
			photo.MessageID = &msgID
		}

		photos = append(photos, photo)
	}

	c.JSON(http.StatusOK, gin.H{
		"ticket":   ticket,
		"messages": messages,
		"photos":   photos,
	})
}

// CreateTicket создает новый тикет
func CreateTicket(c *gin.Context) {
	var request models.NewTicketRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование пользователя
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", request.UserID).Scan(&exists)
	if err != nil {
		logger.LogError("Ошибка при проверке пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании тикета"})
		return
	}

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Если категория не указана, используем значение по умолчанию
	if request.Category == "" {
		request.Category = "спросить"
	}

	// Создаем тикет
	var ticketID int
	err = db.DB.QueryRow(
		"INSERT INTO tickets (user_id, title, description, status, category, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		request.UserID, request.Title, request.Description, "открыт", request.Category, time.Now(),
	).Scan(&ticketID)

	if err != nil {
		logger.LogError("Ошибка при создании тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании тикета"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Тикет создан успешно",
		"ticket_id": ticketID,
	})
}

// UpdateTicket обновляет информацию о тикете
func UpdateTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID тикета"})
		return
	}

	var request models.UpdateTicketRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование тикета
	var exists bool
	var userID int64
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1), user_id FROM tickets WHERE id = $1", id).Scan(&exists, &userID)
	if err != nil {
		logger.LogError("Ошибка при проверке тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении тикета"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Тикет не найден"})
		return
	}

	// Формируем запрос
	query := "UPDATE tickets SET"
	var params []interface{}
	paramCount := 1

	if request.Status != "" {
		query += " status = $" + strconv.Itoa(paramCount)
		params = append(params, request.Status)
		paramCount++

		// Если статус изменился на "закрыт", устанавливаем время закрытия
		if request.Status == "закрыт" {
			if len(params) > 1 {
				query += ","
			}
			query += " closed_at = $" + strconv.Itoa(paramCount)
			params = append(params, time.Now())
			paramCount++
		}
	}

	if request.Category != "" {
		if len(params) > 0 {
			query += ","
		}
		query += " category = $" + strconv.Itoa(paramCount)
		params = append(params, request.Category)
		paramCount++
	}

	// Если нечего обновлять
	if len(params) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нет данных для обновления"})
		return
	}

	query += " WHERE id = $" + strconv.Itoa(paramCount)
	params = append(params, id)

	// Выполняем запрос
	_, err = db.DB.Exec(query, params...)
	if err != nil {
		logger.LogError("Ошибка при обновлении тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении тикета"})
		return
	}

	go func() {
		statusMsg := ""
		if request.Status != "" {
			statusMsg = fmt.Sprintf("Статус вашего тикета %d изменен на '%s'", id, request.Status)
		} else {
			statusMsg = fmt.Sprintf("Ваш тикет %d был обновлен", id)
		}

		data := url.Values{}
		data.Set("super_connect_token", "super_secret_key_2024")
		data.Set("sender_id", "5259653323")
		data.Set("message", statusMsg)
		data.Set("accepter_id", strconv.Itoa(int(userID)))

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

	c.JSON(http.StatusOK, gin.H{
		"message":   "Тикет обновлен успешно",
		"ticket_id": id,
	})
}

// DeleteTicket удаляет тикет
func DeleteTicket(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID тикета"})
		return
	}

	// Проверяем существование тикета
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Ошибка при проверке тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении тикета"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Тикет не найден"})
		return
	}

	// Удаляем связанные сообщения и фотографии
	tx, err := db.DB.Begin()
	if err != nil {
		logger.LogError("Ошибка при создании транзакции: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении тикета"})
		return
	}

	// Удаляем фотографии
	_, err = tx.Exec("DELETE FROM ticket_photos WHERE ticket_id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Ошибка при удалении фотографий тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении тикета"})
		return
	}

	// Удаляем сообщения
	_, err = tx.Exec("DELETE FROM ticket_messages WHERE ticket_id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Ошибка при удалении сообщений тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении тикета"})
		return
	}

	// Удаляем сам тикет
	_, err = tx.Exec("DELETE FROM tickets WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Ошибка при удалении тикета: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении тикета"})
		return
	}

	if err = tx.Commit(); err != nil {
		logger.LogError("Ошибка при фиксации транзакции: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении тикета"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Тикет успешно удален",
		"ticket_id": id,
	})
}
