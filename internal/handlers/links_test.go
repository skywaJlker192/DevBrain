package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тест 1: Неавторизованный доступ
func TestGetLinks_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/links", nil)
	rr := httptest.NewRecorder()

	// Передаём пустой контекст
	GetLinks(nil).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ожидался 401, получен %d", rr.Code)
	}
}

// Тест 2: Некорректный ID для удаления
func TestDeleteLink_InvalidID(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/api/links/abc", nil)
	rr := httptest.NewRecorder()

	DeleteLink(nil).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("ожидался 400, получен %d", rr.Code)
	}
}

// Тест 3: Экспорт без токена
func TestExport_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/export?format=json", nil)
	rr := httptest.NewRecorder()

	ExportData(nil).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ожидался 401, получен %d", rr.Code)
	}
}
