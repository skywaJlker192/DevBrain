package handlers

import (
	"encoding/json"
	"net/http"

	"devbrain-pro/internal/config"
	"devbrain-pro/internal/models"
	"devbrain-pro/internal/security"
)

type AuthService struct {
	cfg    config.Config
	hasher *security.PasswordHasher
	jwt    *security.JWTManager
	repo   UserRepository
}

type UserRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(*models.User) error
}

func NewAuthService(cfg config.Config, hasher *security.PasswordHasher, jwt *security.JWTManager, repo UserRepository) *AuthService {
	return &AuthService{cfg: cfg, hasher: hasher, jwt: jwt, repo: repo}
}

// Register: Регистрация пользователя (Требование #1, #2, #3)
func (s *AuthService) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	// Требование #2: Корректная обработка ошибок
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный формат запроса",
		})
		return
	}

	// Требование #1: Белая валидация входных данных
	v := security.NewValidator()
	if err := v.ValidateEmail(req.Email); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	if err := v.ValidatePassword(req.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Проверка: пользователь уже существует
	if existing, _ := s.repo.GetUserByEmail(req.Email); existing != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Пользователь уже существует",
		})
		return
	}

	// Требование #3: Хеширование пароля через Argon2 (≥100мс на подбор)
	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Ошибка сервера",
		})
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: hash,
		Name:     req.Name,
		Role:     "user",
	}

	if err := s.repo.CreateUser(user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Ошибка сохранения",
		})
		return
	}

	// Требование #2: Клиенту - общее сообщение без деталей
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Регистрация успешна",
	})
}

// Login: Аутентификация с JWT (Требование #3)
func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный формат запроса",
		})
		return
	}

	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		// Требование #2: Не раскрываем существование пользователя
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный email или пароль",
		})
		return
	}

	// Требование #3: Constant-time сравнение паролей (Argon2)
	if ok, err := s.hasher.Compare(req.Password, user.Password); err != nil || !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Неверный email или пароль",
		})
		return
	}

	// Требование #3: Генерация JWT с коротким TTL (30 мин)
	token, err := s.jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Ошибка сервера",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{
		User:        *user,
		AccessToken: token,
	})
}
