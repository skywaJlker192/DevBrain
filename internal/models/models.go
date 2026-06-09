package models

import (
	"time"
)

// User модель
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Никогда не возвращаем
	Name      string    `json:"name"`
	Role      string    `json:"role"` // "user" | "admin"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Link модель
type Link struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	IsArchived  bool      `json:"is_archived"`
	ReadingTime int       `json:"reading_time"` // минуты
	MetaData    MetaData  `json:"meta_data,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MetaData struct {
	SiteName    string `json:"site_name,omitempty"`
	Author      string `json:"author,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	Favicon     string `json:"favicon,omitempty"`
}

// Stats для аналитики
type UserStats struct {
	TotalLinks      int64   `json:"total_links"`
	UniqueTags      int64   `json:"unique_tags"`
	TotalReadingTime int64  `json:"total_reading_time"` // ✅ ДОБАВЛЕНО
	ArchivedLinks   int64   `json:"archived_links"`
	TopTags         []TagCount `json:"top_tags"`
}

type TagCount struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

type DailyStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// DTO для создания/обновления
type CreateLinkRequest struct {
	URL  string   `json:"url"`
	Tags []string `json:"tags"`
}

type UpdateLinkRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsArchived  *bool    `json:"is_archived,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
