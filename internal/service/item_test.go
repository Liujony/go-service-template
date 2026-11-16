package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-service-template/internal/model"
	"go-service-template/internal/repository"
)

type stubStore struct {
	created  *model.Item
	existing model.Item
	missing  bool
}

func (s *stubStore) Create(_ context.Context, item *model.Item) error {
	item.ID = 1
	s.created = item
	return nil
}

func (s *stubStore) Get(_ context.Context, id uint) (model.Item, error) {
	if s.missing {
		return model.Item{}, repository.ErrNotFound
	}
	item := s.existing
	item.ID = id
	return item, nil
}

func (s *stubStore) List(context.Context, int, int) ([]model.Item, int64, error) {
	return nil, 0, nil
}

func (s *stubStore) Update(_ context.Context, item *model.Item) error {
	s.existing = *item
	return nil
}

func (s *stubStore) Delete(_ context.Context, _ uint) error { return nil }

func TestCreateAppliesDefaultsAndTrims(t *testing.T) {
	store := &stubStore{}
	svc := NewItems(store)

	item, err := svc.Create(context.Background(), CreateInput{
		Name:        "  Ship it  ",
		Description: "  with tests  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Ship it" || item.Description != "with tests" {
		t.Fatalf("values were not trimmed: %+v", item)
	}
	if item.Status != model.StatusActive {
		t.Fatalf("expected default status active, got %q", item.Status)
	}
	if item.ID != 1 {
		t.Fatalf("store did not assign an id: %+v", item)
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	svc := NewItems(&stubStore{})
	cases := []struct {
		name  string
		input CreateInput
		field string
	}{
		{"missing name", CreateInput{}, "name"},
		{"bad status", CreateInput{Name: "ok", Status: "deleted"}, "status"},
		{"long name", CreateInput{Name: strings.Repeat("a", 121)}, "name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.input)
			var validation *ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if _, ok := validation.AsFields()[tc.field]; !ok {
				t.Fatalf("expected field %q in %v", tc.field, validation.AsFields())
			}
		})
	}
}

func TestUpdateOnlyChangesProvidedFields(t *testing.T) {
	store := &stubStore{existing: model.Item{Name: "old", Description: "keep me", Status: model.StatusActive}}
	svc := NewItems(store)

	newName := "new"
	item, err := svc.Update(context.Background(), 7, UpdateInput{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "new" || item.Description != "keep me" {
		t.Fatalf("unexpected partial update: %+v", item)
	}
}

func TestUpdateValidatesTheMergedResult(t *testing.T) {
	store := &stubStore{existing: model.Item{Name: "old", Status: model.StatusActive}}
	svc := NewItems(store)
	empty := "  "

	_, err := svc.Update(context.Background(), 3, UpdateInput{Name: &empty})
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestUpdatePropagatesNotFound(t *testing.T) {
	svc := NewItems(&stubStore{missing: true})
	if _, err := svc.Update(context.Background(), 9, UpdateInput{}); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListClampsPagination(t *testing.T) {
	store := &stubStore{}
	svc := NewItems(store)
	if _, _, err := svc.List(context.Background(), 10_000, -5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
