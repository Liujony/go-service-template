// Package handler translates HTTP to service calls and back.
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-service-template/internal/model"
	"go-service-template/internal/repository"
	"go-service-template/internal/service"
	"go-service-template/pkg/response"
)

// ItemHandler owns the /api/v1/items routes.
type ItemHandler struct {
	service service.ItemService
	log     *slog.Logger
}

// NewItemHandler wires the handler.
func NewItemHandler(items service.ItemService, log *slog.Logger) *ItemHandler {
	return &ItemHandler{service: items, log: log}
}

type createRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type updateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

// Create handles POST /api/v1/items.
func (h *ItemHandler) Create(c *gin.Context) {
	var body createRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_json", "request body must be a JSON object")
		return
	}
	item, err := h.service.Create(c.Request.Context(), service.CreateInput{
		Name:        body.Name,
		Description: body.Description,
		Status:      body.Status,
	})
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.Created(c, item)
}

// List handles GET /api/v1/items?limit=&offset=.
func (h *ItemHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.OK(c, http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

// Get handles GET /api/v1/items/:id.
func (h *ItemHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.OK(c, http.StatusOK, item)
}

// Update handles PUT /api/v1/items/:id. Only the fields present in the body change.
func (h *ItemHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body updateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_json", "request body must be a JSON object")
		return
	}
	item, err := h.service.Update(c.Request.Context(), id, service.UpdateInput{
		Name:        body.Name,
		Description: body.Description,
		Status:      body.Status,
	})
	if err != nil {
		h.respondError(c, err)
		return
	}
	response.OK(c, http.StatusOK, item)
}

// Delete handles DELETE /api/v1/items/:id.
func (h *ItemHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ItemHandler) respondError(c *gin.Context, err error) {
	var validation *service.ValidationError
	switch {
	case errors.As(err, &validation):
		response.Invalid(c, validation.AsFields())
	case errors.Is(err, repository.ErrNotFound):
		response.Fail(c, http.StatusNotFound, "not_found", "item not found")
	default:
		h.log.Error("request failed",
			"path", c.FullPath(),
			"request_id", c.GetString("request_id"),
			"error", err,
		)
		response.Fail(c, http.StatusInternalServerError, "internal_error", "something went wrong")
	}
}

func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid_id", "id must be a positive integer")
		return 0, false
	}
	return uint(value), true
}

// ensure the compile-time contract is visible to readers
var _ = model.StatusActive
