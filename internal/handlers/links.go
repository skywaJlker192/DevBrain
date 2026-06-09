package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"net/http"
	"strconv"
	"strings"

	"devbrain-pro/internal/models"
	"devbrain-pro/internal/repository"
	"devbrain-pro/internal/security"
	"github.com/go-chi/chi/v5"
)

func GetLinks(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		search := strings.TrimSpace(r.URL.Query().Get("search"))
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 || limit > 100 {
			limit = 50
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		// Фильтрация по тегам (полное совпадение)
		tagsQuery := r.URL.Query().Get("tags")
		var tagFilters []string
		if tagsQuery != "" {
			tagFilters = strings.Split(tagsQuery, ",")
			// Очистка тегов от пробелов
			for i, tag := range tagFilters {
				tagFilters[i] = strings.TrimSpace(tag)
			}
		}

		links, err := repo.GetLinks(user.ID, repository.LinkFilters{
			Search:   search,
			Tags:     tagFilters,
			Archived: false,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			// ✅ ВАЖНО: Возвращаем JSON ошибку, а не текст
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Ошибка получения данных",
				"message": err.Error(), // Можно отправить детальную ошибку в консоль
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"links": links, "total": len(links)})
	}
}
func CreateLink(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			URL  string   `json:"url"`
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Неверный формат запроса",
			})
			return
		}

		if err := security.NewValidator().ValidateURL(req.URL); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		title := extractTitleFromURL(req.URL)

		link := &models.Link{
			UserID:     user.ID,
			URL:        req.URL,
			Title:      title,
			Tags:       req.Tags,
			IsArchived: false,
		}

		if err := repo.CreateLink(link); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Ошибка сохранения",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(link)
	}
}
func DeleteLink(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			http.Error(w, "ID не указан", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}

		if err := repo.DeleteLink(id, user.ID); err != nil {
			http.Error(w, "Ошибка удаления или доступ запрещён", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func GetStats(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(UserContextKey).(*models.User)
		stats, err := repo.GetUserStats(user.ID)
		if err != nil {
			http.Error(w, "Ошибка статистики", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

func ExportData(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(UserContextKey).(*models.User)
		format := r.URL.Query().Get("format")
		if format != "csv" {
			format = "json"
		}

		links, err := repo.GetLinks(user.ID, repository.LinkFilters{Archived: false, Limit: 5000})
		if err != nil {
			http.Error(w, "Ошибка экспорта", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=devbrain-%s.%s", user.Email, format))

		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			cw := csv.NewWriter(w)
			cw.Write([]string{"ID", "URL", "Title", "Tags", "Created"})
			for _, l := range links {
				cw.Write([]string{
					strconv.FormatInt(l.ID, 10),
					l.URL,
					l.Title,
					strings.Join(l.Tags, ","),
					l.CreatedAt.Format("2006-01-02"),
				})
			}
			cw.Flush()
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(links)
		}
	}
}

func GetLink(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

func UpdateLink(repo *repository.PostgresRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

func extractTitleFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "Без названия"
	}

	host := parsed.Hostname()
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	switch {
	case strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be"):
		return "YouTube Video"
	case strings.Contains(host, "github.com"):
		return "GitHub Repository"
	case strings.Contains(host, "twitter.com") || strings.Contains(host, "x.com"):
		return "X (Twitter) Post"
	case strings.Contains(host, "habr.com") || strings.Contains(host, "habr.ru"):
		return "Habr Article"
	}

	return host
}
