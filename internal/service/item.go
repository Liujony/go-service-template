// Package service holds the business rules. It has no HTTP types and no SQL.
package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"go-service-template/internal/model"
	"go-service-template/internal/repository"
)

// FieldError carries a validation failure for one request field.
type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string { return e.Field + ": " + e.Message }

// ValidationError groups one or more field errors.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		parts = append(parts, field.Error())
	}
	return strings.Join(parts, "; ")
}

// AsFields flattens the error for the API envelope.
func (e *ValidationError) AsFields() map[string]string {
	out := make(map[string]string, len(e.Fields))
	for _, field := range e.Fields {
		out[field.Field] = field.Message
	}
	return out
}

// CreateInput is the service's view of a create request.
type CreateInput struct {
	Name        string
	Description string
	Status      string
}

// UpdateInput uses pointers so "not provided" and "empty string" stay distinct.
type UpdateInput struct {
	Name        *string
	Description *string
	Status      *string
}

// ItemService is the API-independent contract.
type ItemService interface {
	Create(ctx context.Context, input CreateInput) (model.Item, error)
	Get(ctx context.Context, id uint) (model.Item, error)
	List(ctx context.Context, limit, offset int) ([]model.Item, int64, error)
	Update(ctx context.Context, id uint, input UpdateInput) (model.Item, error)
	Delete(ctx context.Context, id uint) error
}

// Items implements ItemService on top of a store.
type Items struct {
	store repository.ItemStore
}

// NewItems builds the service.
func NewItems(store repository.ItemStore) *Items {
	return &Items{store: store}
}

// Create validates and persists a new item.
func (s *Items) Create(ctx context.Context, input CreateInput) (model.Item, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = model.StatusActive
	}
	if err := validate(name, description, status); err != nil {
		return model.Item{}, err
	}
	item := model.Item{Name: name, Description: description, Status: status}
	if err := s.store.Create(ctx, &item); err != nil {
		return model.Item{}, err
	}
	return item, nil
}

// Get returns one item or repository.ErrNotFound.
func (s *Items) Get(ctx context.Context, id uint) (model.Item, error) {
	return s.store.Get(ctx, id)
}

// List clamps pagination to something a database can survive.
func (s *Items) List(ctx context.Context, limit, offset int) ([]model.Item, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.store.List(ctx, limit, offset)
}

// Update applies only the provided fields, then validates the merged result.
func (s *Items) Update(ctx context.Context, id uint, input UpdateInput) (model.Item, error) {
	item, err := s.store.Get(ctx, id)
	if err != nil {
		return model.Item{}, err
	}
	if input.Name != nil {
		item.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		item.Status = strings.TrimSpace(*input.Status)
	}
	if err := validate(item.Name, item.Description, item.Status); err != nil {
		return model.Item{}, err
	}
	if err := s.store.Update(ctx, &item); err != nil {
		return model.Item{}, err
	}
	return item, nil
}

// Delete removes an item or reports that it never existed.
func (s *Items) Delete(ctx context.Context, id uint) error {
	return s.store.Delete(ctx, id)
}

func validate(name, description, status string) error {
	var fields []FieldError
	if name == "" {
		fields = append(fields, FieldError{Field: "name", Message: "is required"})
	} else if utf8.RuneCountInString(name) > 120 {
		fields = append(fields, FieldError{Field: "name", Message: "must be at most 120 characters"})
	}
	if utf8.RuneCountInString(description) > 2000 {
		fields = append(fields, FieldError{Field: "description", Message: "must be at most 2000 characters"})
	}
	if !model.ValidStatus(status) {
		fields = append(fields, FieldError{Field: "status", Message: "must be active or archived"})
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
