package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"go-service-template/internal/database"
	"go-service-template/internal/handler"
	"go-service-template/internal/repository"
	"go-service-template/internal/router"
	"go-service-template/internal/service"
	"go-service-template/pkg/logger"
)

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("database: %v", err)
	}
	store := repository.NewItemStore(db)
	items := service.NewItems(store)
	engine := router.New(handler.NewItemHandler(items, logger.New("ERROR")), logger.New("ERROR"), "test")
	return engine
}

func do(t *testing.T, engine *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload *bytes.Reader
	if body != nil {
		encoded, _ := json.Marshal(body)
		payload = bytes.NewReader(encoded)
	} else {
		payload = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, path, payload)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func decode(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode %q: %v", recorder.Body.String(), err)
	}
	return envelope
}

func TestCreateListUpdateDelete(t *testing.T) {
	engine := newEngine(t)

	created := do(t, engine, http.MethodPost, "/api/v1/items", map[string]string{
		"name": "Write the template", "description": "small and complete",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", created.Code, created.Body.String())
	}
	if created.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID header")
	}
	data := decode(t, created)["data"].(map[string]any)
	id := int(data["id"].(float64))
	if data["status"] != "active" {
		t.Fatalf("expected default status active, got %v", data["status"])
	}

	list := do(t, engine, http.MethodGet, "/api/v1/items", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d", list.Code)
	}
	listData := decode(t, list)["data"].(map[string]any)
	if int(listData["total"].(float64)) != 1 {
		t.Fatalf("expected total 1, got %v", listData["total"])
	}

	updated := do(t, engine, http.MethodPut, "/api/v1/items/"+strconv.Itoa(id), map[string]string{"status": "archived"})
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d body = %s", updated.Code, updated.Body.String())
	}
	if decode(t, updated)["data"].(map[string]any)["status"] != "archived" {
		t.Fatal("status was not updated")
	}

	deleted := do(t, engine, http.MethodDelete, "/api/v1/items/"+strconv.Itoa(id), nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", deleted.Code)
	}
	missing := do(t, engine, http.MethodGet, "/api/v1/items/"+strconv.Itoa(id), nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", missing.Code)
	}
}

func TestValidationReturnsFieldErrors(t *testing.T) {
	engine := newEngine(t)
	response := do(t, engine, http.MethodPost, "/api/v1/items", map[string]string{"status": "nope"})
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	fields := decode(t, response)["error"].(map[string]any)["fields"].(map[string]any)
	if _, ok := fields["name"]; !ok {
		t.Fatalf("expected name error, got %v", fields)
	}
	if _, ok := fields["status"]; !ok {
		t.Fatalf("expected status error, got %v", fields)
	}
}

func TestMalformedJSONAndBadID(t *testing.T) {
	engine := newEngine(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader([]byte("{nope")))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed json status = %d", recorder.Code)
	}

	badID := do(t, engine, http.MethodGet, "/api/v1/items/abc", nil)
	if badID.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d", badID.Code)
	}
}

func TestHealthAndHome(t *testing.T) {
	engine := newEngine(t)

	health := do(t, engine, http.MethodGet, "/healthz", nil)
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d", health.Code)
	}

	home := do(t, engine, http.MethodGet, "/", nil)
	if home.Code != http.StatusOK {
		t.Fatalf("home status = %d", home.Code)
	}
	if !bytes.Contains(home.Body.Bytes(), []byte("go-service-template")) {
		t.Fatal("home page does not look like the dashboard")
	}
}
