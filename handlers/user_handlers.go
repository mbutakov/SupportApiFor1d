package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"support_front_api/db"
	"support_front_api/logger"
	"support_front_api/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetUsers возвращает список пользователей
func GetUsers(c *gin.Context) {
	// Пагинация
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Получаем пользователей из базы данных
	rows, err := db.DB.Query(
		"SELECT id, full_name, phone, location_lat, location_lng, birth_date, is_registered, registered_at FROM users ORDER BY id LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		logger.LogError("Ошибка при получении пользователей: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении пользователей"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var birthDate sql.NullTime
		var registeredAt sql.NullTime

		if err := rows.Scan(
			&user.ID,
			&user.FullName,
			&user.Phone,
			&user.LocationLat,
			&user.LocationLng,
			&birthDate,
			&user.IsRegistered,
			&registeredAt,
		); err != nil {
			logger.LogError("Ошибка при сканировании пользователя: %v", err)
			continue
		}

		if birthDate.Valid {
			user.BirthDate = birthDate.Time
		}

		if registeredAt.Valid {
			regAt := registeredAt.Time
			user.RegisteredAt = &regAt
		}

		users = append(users, user)
	}

	// Получаем общее количество пользователей
	var count int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		logger.LogError("Ошибка при подсчете пользователей: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": count,
		"page":  page,
		"limit": limit,
	})
}

// GetUserById возвращает пользователя по ID
func GetUserById(c *gin.Context) {
	// Получаем ID пользователя
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	// Получаем пользователя из базы данных
	var user models.User
	var birthDate sql.NullTime
	var registeredAt sql.NullTime

	err = db.DB.QueryRow(
		"SELECT id, full_name, phone, location_lat, location_lng, birth_date, is_registered, registered_at FROM users WHERE id = $1",
		userID,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Phone,
		&user.LocationLat,
		&user.LocationLng,
		&birthDate,
		&user.IsRegistered,
		&registeredAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		} else {
			logger.LogError("Ошибка при получении пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении пользователя"})
		}
		return
	}

	if birthDate.Valid {
		user.BirthDate = birthDate.Time
	}

	if registeredAt.Valid {
		regAt := registeredAt.Time
		user.RegisteredAt = &regAt
	}

	// Получаем тикеты пользователя
	rows, err := db.DB.Query(
		"SELECT id, user_id, title, description, status, category, created_at, closed_at FROM tickets WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		logger.LogError("Ошибка при получении тикетов пользователя: %v", err)
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
			logger.LogError("Ошибка при сканировании тикета: %v", err)
			continue
		}

		if closedAt.Valid {
			closedAtTime := closedAt.Time
			ticket.ClosedAt = &closedAtTime
		}

		tickets = append(tickets, ticket)
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"tickets": tickets,
	})
}

// CreateUser создает нового пользователя
func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь с таким ID
	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID).Scan(&exists); err != nil {
		logger.LogError("Ошибка при проверке существования пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании пользователя"})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь с таким ID уже существует"})
		return
	}

	// Устанавливаем значения по умолчанию
	if !user.IsRegistered {
		user.RegisteredAt = nil
	} else if user.RegisteredAt == nil {
		now := time.Now()
		user.RegisteredAt = &now
	}

	// Создаем пользователя
	_, err := db.DB.Exec(
		"INSERT INTO users (id, full_name, phone, location_lat, location_lng, birth_date, is_registered, registered_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		user.ID,
		user.FullName,
		user.Phone,
		user.LocationLat,
		user.LocationLng,
		user.BirthDate,
		user.IsRegistered,
		user.RegisteredAt,
	)

	if err != nil {
		logger.LogError("Ошибка при создании пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании пользователя"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно создан",
		"user_id": user.ID,
	})
}

// UpdateUser обновляет информацию о пользователе
func UpdateUser(c *gin.Context) {
	// Получаем ID пользователя
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	// Проверяем существование пользователя
	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists); err != nil {
		logger.LogError("Ошибка при проверке существования пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользователя"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Получаем текущие данные пользователя
	var currentUser models.User
	var birthDate sql.NullTime
	var registeredAt sql.NullTime

	err = db.DB.QueryRow(
		"SELECT id, full_name, phone, location_lat, location_lng, birth_date, is_registered, registered_at FROM users WHERE id = $1",
		userID,
	).Scan(
		&currentUser.ID,
		&currentUser.FullName,
		&currentUser.Phone,
		&currentUser.LocationLat,
		&currentUser.LocationLng,
		&birthDate,
		&currentUser.IsRegistered,
		&registeredAt,
	)

	if err != nil {
		logger.LogError("Ошибка при получении текущих данных пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользователя"})
		return
	}

	if birthDate.Valid {
		currentUser.BirthDate = birthDate.Time
	}

	if registeredAt.Valid {
		regAt := registeredAt.Time
		currentUser.RegisteredAt = &regAt
	}

	// Получаем новые данные
	var updateUser models.User
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только те поля, которые были указаны
	if updateUser.FullName != "" {
		currentUser.FullName = updateUser.FullName
	}

	if updateUser.Phone != "" {
		currentUser.Phone = updateUser.Phone
	}

	// Для координат нужно проверять, были ли они переданы
	if updateUser.LocationLat != 0 {
		currentUser.LocationLat = updateUser.LocationLat
	}

	if updateUser.LocationLng != 0 {
		currentUser.LocationLng = updateUser.LocationLng
	}

	// Если пользователь зарегистрировался
	if updateUser.IsRegistered && !currentUser.IsRegistered {
		currentUser.IsRegistered = true
		now := time.Now()
		currentUser.RegisteredAt = &now
	}

	// Обновляем данные пользователя
	_, err = db.DB.Exec(
		"UPDATE users SET full_name = $1, phone = $2, location_lat = $3, location_lng = $4, birth_date = $5, is_registered = $6, registered_at = $7 WHERE id = $8",
		currentUser.FullName,
		currentUser.Phone,
		currentUser.LocationLat,
		currentUser.LocationLng,
		currentUser.BirthDate,
		currentUser.IsRegistered,
		currentUser.RegisteredAt,
		userID,
	)

	if err != nil {
		logger.LogError("Ошибка при обновлении пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользователя"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Пользователь успешно обновлен",
		"user_id": userID,
	})
}
