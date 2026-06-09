package repository

import (
	"database/sql"
	"devbrain-pro/internal/models"
	"encoding/json"
	"fmt"
	"strings" // ← ДОБАВЛЕНО

)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

type LinkFilters struct {
	Search   string
	Tags     []string
	Archived bool
	Limit    int
	Offset   int
}

// GetUserByEmail: Поиск пользователя по email (Требование #4)
func (r *PostgresRepo) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, created_at, updated_at
		FROM users
		WHERE email = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	return user, nil
}

// CreateUser: Создание пользователя (Требование #4)
func (r *PostgresRepo) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.Exec(query,
		user.Email, user.Password, user.Name, user.Role,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user id: %w", err)
	}

	user.ID = id
	return nil
}

// GetLinks: Поиск по title ИЛИ тегам (для SQLite)
func (r *PostgresRepo) GetLinks(userID int64, f LinkFilters) ([]models.Link, error) {
	query := `
		SELECT id, user_id, url, title, description, tags, is_archived,
		       reading_time, created_at, updated_at
		FROM links
		WHERE user_id = ? AND is_archived = ?`

	args := []interface{}{userID, f.Archived}

	// 🔍 Поиск по заголовку (title)
	if f.Search != "" {
		query += ` AND LOWER(title) LIKE ?`
		searchTerm := "%" + strings.ToLower(f.Search) + "%"
		args = append(args, searchTerm)
	}

	// ️ Фильтрация по тегам (для SQLite — перебираем все ссылки и проверяем теги вручную)
	if len(f.Tags) > 0 {
		// Получаем все link IDs для текущего пользователя
		linkIDs := make([]int64, 0)

		rows, err := r.db.Query(`
			SELECT id FROM links WHERE user_id = ? AND is_archived = 0
		`, userID)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err == nil {
				linkIDs = append(linkIDs, id)
			}
		}
		rows.Close()

		// Теперь проверяем каждый тег
		for _, tag := range f.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}

			// Ищем ссылки, содержащие этот тег
			var matchingIDs []int64
			rows, err := r.db.Query(`
				SELECT id FROM links WHERE id IN (?) AND tags LIKE ?
			`, linkIDs, "%"+tag+"%")
			if err != nil {
				continue
			}

			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err == nil {
					matchingIDs = append(matchingIDs, id)
				}
			}
			rows.Close()

			// Объединяем результаты
			if len(matchingIDs) > 0 {
				linkIDs = intersect(linkIDs, matchingIDs)
			}
		}

		// Формируем новый запрос с ID ссылок
		if len(linkIDs) > 0 {
			idList := make([]interface{}, len(linkIDs))
			for i, id := range linkIDs {
				idList[i] = id
			}
			query += ` AND id IN (` + strings.Repeat(",?", len(linkIDs)-1) + `)`
			args = append(args, idList...)
		} else {
			// Нет совпадений — возвращаем пустой список
			return []models.Link{}, nil
		}
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	var links []models.Link
	for rows.Next() {
		var l models.Link
		var tagsJSON string
		err := rows.Scan(
			&l.ID, &l.UserID, &l.URL, &l.Title, &l.Description,
			&tagsJSON, &l.IsArchived, &l.ReadingTime,
			&l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		if tagsJSON != "" && tagsJSON != "[]" {
			json.Unmarshal([]byte(tagsJSON), &l.Tags)
		} else {
			l.Tags = []string{}
		}
		links = append(links, l)
	}
	return links, nil
}

// helper function to intersect two slices
func intersect(a, b []int64) []int64 {
	set := make(map[int64]bool)
	result := []int64{}

	for _, v := range a {
		set[v] = true
	}

	for _, v := range b {
		if set[v] {
			result = append(result, v)
		}
	}

	return result
}
// CreateLink: Создание ссылки (Требование #4)
func (r *PostgresRepo) CreateLink(l *models.Link) error {
	tj, _ := json.Marshal(l.Tags)

	query := `
		INSERT INTO links (user_id, url, title, description, tags, is_archived, reading_time, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.Exec(query,
		l.UserID, l.URL, l.Title, l.Description, string(tj), l.IsArchived, l.ReadingTime,
	)
	if err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	id, _ := result.LastInsertId()
	l.ID = id
	return nil
}

// DeleteLink: Удаление ссылки (Требование #3)
func (r *PostgresRepo) DeleteLink(id, userID int64) error {
	query := `DELETE FROM links WHERE id = ? AND user_id = ?`

	res, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("link not found or access denied")
	}

	return nil
}

// GetUserStats: Статистика пользователя (Требование #3)
func (r *PostgresRepo) GetUserStats(userID int64) (*models.UserStats, error) {
	s := &models.UserStats{}

	// Всего ссылок
	err := r.db.QueryRow("SELECT COUNT(*) FROM links WHERE user_id = ?", userID).Scan(&s.TotalLinks)
	if err != nil {
		return nil, fmt.Errorf("count links failed: %w", err)
	}

	// Уникальные теги (через jsonb_array_elements_text для PostgreSQL)
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT t.tag)
		FROM links, jsonb_array_elements_text(tags::jsonb) AS t
		WHERE user_id = ?`, userID).Scan(&s.UniqueTags)
	if err != nil {
		s.UniqueTags = 0 // Если ошибка — ставим 0
	}

	// Время чтения (сумма reading_time)
	err = r.db.QueryRow("SELECT COALESCE(SUM(reading_time), 0) FROM links WHERE user_id = ?", userID).Scan(&s.TotalReadingTime)
	if err != nil {
		s.TotalReadingTime = 0
	}

	return s, nil
}
