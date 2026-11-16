package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go-service-template/pkg/response"
)

func newContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	return context, recorder
}

func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return envelope
}

func TestOKAndCreated(t *testing.T) {
	context, recorder := newContext()
	response.OK(context, http.StatusOK, gin.H{"ok": true})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	if decodeEnvelope(t, recorder)["data"].(map[string]any)["ok"] != true {
		t.Fatal("data missing")
	}

	context, recorder = newContext()
	response.Created(context, gin.H{"id": 1})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestFailCarriesACode(t *testing.T) {
	context, recorder := newContext()
	response.Fail(context, http.StatusNotFound, "not_found", "item not found")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d", recorder.Code)
	}
	errBody := decodeEnvelope(t, recorder)["error"].(map[string]any)
	if errBody["code"] != "not_found" || errBody["message"] != "item not found" {
		t.Fatalf("unexpected error envelope: %v", errBody)
	}
	if !context.IsAborted() {
		t.Fatal("Fail must abort the handler chain")
	}
}

func TestInvalidListsFields(t *testing.T) {
	context, recorder := newContext()
	response.Invalid(context, map[string]string{"name": "is required"})
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", recorder.Code)
	}
	fields := decodeEnvelope(t, recorder)["error"].(map[string]any)["fields"].(map[string]any)
	if fields["name"] != "is required" {
		t.Fatalf("unexpected fields: %v", fields)
	}
}
