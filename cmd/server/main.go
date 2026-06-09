package main

import (
	"database/sql"
	"devbrain-pro/internal/config"
	"devbrain-pro/internal/handlers"
	"devbrain-pro/internal/repository"
	"devbrain-pro/internal/security"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite" // pure-Go SQLite драйвер, не требует CGO
)

func main() {
	// Load configuration (Требование #5: секреты из env)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database (Требование #4: prepared statements)
	db, err := connectDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	linkRepo := repository.NewPostgresRepo(db)

	// ============================================
	// ИНИЦИАЛИЗАЦИЯ БЕЗОПАСНОСТИ (Требование #3)
	// ============================================
	// Хеширование паролей: Argon2id (Требование #3: ≥100мс на подбор)
	hasher := security.NewPasswordHasher(
		cfg.Security.Argon2Memory,
		cfg.Security.Argon2Iterations,
		cfg.Security.Argon2SaltLen,
		cfg.Security.Argon2KeyLen,
	)

	// JWT менеджер: короткий TTL, подпись из env (Требование #3, #5)
	jwtMgr := security.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.AccessTTL,
	)

	// AuthService с реальными зависимостями
	authService := handlers.NewAuthService(*cfg, hasher, jwtMgr, linkRepo)

	// Initialize middleware (Требование #6: rate limiting, CORS, security headers)
	mw := handlers.NewMiddleware(*cfg)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mw.Logger)          // Требование #2: логирование без утечек
	r.Use(mw.Recoverer)       // Защита от паник
	r.Use(mw.SecurityHeaders) // Требование #6: CSP, X-Frame-Options, etc.
	r.Use(mw.CORS(cfg.Server.AllowedOrigins)) // Требование #6: явные домены
	r.Use(mw.RequestSizeLimit(cfg.Server.MaxBodySize)) // Требование #6: защита от DoS

	// ============================================
	// 1. СТАТИЧЕСКИЕ ФАЙЛЫ (строго ПЕРЕД catch-all)
	// ============================================
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	staticDir := filepath.Join(baseDir, "static")

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// ============================================
	// 2. API МАРШРУТЫ (public) — АУТЕНТИФИКАЦИЯ
	// Требование #3: реальная генерация JWT
	// ============================================
	r.Post("/api/auth/register", authService.Register)
	r.Post("/api/auth/login", authService.Login)

	// ============================================
	// 3. API МАРШРУТЫ (protected)
	// Требование #3: проверка токена ПЕРЕД логикой
	// Требование #4: принцип наименьших привилегий
	// ============================================
	r.Group(func(r chi.Router) {
		r.Use(mw.RateLimit) // Требование #6: не более 10 запросов/мин
		r.Use(mw.Auth)      // Требование #3: проверка реального JWT

		r.Get("/api/links", handlers.GetLinks(linkRepo))
		r.Post("/api/links", handlers.CreateLink(linkRepo))
		r.Get("/api/links/{id}", handlers.GetLink(linkRepo))
		r.Put("/api/links/{id}", handlers.UpdateLink(linkRepo))
		r.Delete("/api/links/{id}", handlers.DeleteLink(linkRepo))
		r.Get("/api/stats", handlers.GetStats(linkRepo))
		r.Get("/api/export", handlers.ExportData(linkRepo))
	})

	// ============================================
	// 4. HEALTH CHECK
	// ============================================
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// ============================================
	// 5. SPA FALLBACK (строго В КОНЦЕ!)
	// ============================================
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 Server starting on http://localhost:8080/")
	log.Printf("📊 Environment: %s", cfg.Server.Env)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// connectDB: подключение к БД (Требование #4: prepared statements в repository/)
func connectDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	// Для локальной разработки: SQLite
	// Для продакшена: раскомментируй блок PostgreSQL ниже

	// === SQLite для локального теста ===
	db, err := sql.Open("sqlite", "./devbrain.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err) // Требование #2: %w для контекста
	}

	// Проверка подключения (Требование #2: логирование без утечек)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	// SQLite: одно соединение
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil

	// === PostgreSQL для продакшена ===
	/*
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
	*/
}

// runMigrations: создание таблиц (Требование #4: параметризация в repository/)
func runMigrations(db *sql.DB) error {
	// Включаем поддержку внешних ключей в SQLite
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			title TEXT,
			description TEXT,
			tags TEXT NOT NULL DEFAULT '[]',
			is_archived INTEGER NOT NULL DEFAULT 0,
			reading_time INTEGER NOT NULL DEFAULT 0,
			meta_data TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_links_created_at ON links(created_at DESC)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err) // Требование #2
		}
	}

	log.Println("✅ Migrations completed")
	return nil
}
