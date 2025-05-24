package models

import "time"

// User представляет модель пользователя
type User struct {
	ID           int64      `json:"id"`
	FullName     string     `json:"full_name"`
	Phone        string     `json:"phone"`
	LocationLat  float64    `json:"location_lat"`
	LocationLng  float64    `json:"location_lng"`
	BirthDate    time.Time  `json:"birth_date"`
	IsRegistered bool       `json:"is_registered"`
	RegisteredAt *time.Time `json:"registered_at"`
}

// Ticket представляет модель тикета поддержки
type Ticket struct {
	ID          int        `json:"id"`
	UserID      int64      `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Category    string     `json:"category"`
	CreatedAt   time.Time  `json:"created_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

// TicketMessage представляет модель сообщения в тикете
type TicketMessage struct {
	ID         int       `json:"id"`
	TicketID   int       `json:"ticket_id"`
	SenderType string    `json:"sender_type"` // 'user' или 'support'
	SenderID   int64     `json:"sender_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

// TicketPhoto представляет модель фотографии в тикете
type TicketPhoto struct {
	ID         int       `json:"id"`
	TicketID   int       `json:"ticket_id"`
	SenderType string    `json:"sender_type"` // 'user' или 'support'
	SenderID   int64     `json:"sender_id"`
	FilePath   string    `json:"file_path"`
	FileID     string    `json:"file_id"`
	MessageID  *int      `json:"message_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewTicketRequest представляет запрос на создание нового тикета
type NewTicketRequest struct {
	UserID      int64  `json:"user_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category"`
}

// UpdateTicketRequest представляет запрос на обновление тикета
type UpdateTicketRequest struct {
	Status   string `json:"status"`
	Category string `json:"category"`
}

// NewMessageRequest представляет запрос на создание нового сообщения
type NewMessageRequest struct {
	SenderType string `json:"sender_type" binding:"required"`
	SenderID   int64  `json:"sender_id" binding:"required"`
	Message    string `json:"message" binding:"required"`
}
