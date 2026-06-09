package handlers

import "devbrain-pro/internal/models"

// Требование #3: Типизированный ключ контекста (защита от коллизий)
type contextKey string

const UserContextKey contextKey = "user"

// GetUserFromContext безопасно извлекает пользователя
func GetUserFromContext(ctx interface{}) (*models.User, bool) {
	user, ok := ctx.(*models.User)
	return user, ok
}
