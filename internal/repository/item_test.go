package repository_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"go-service-template/internal/database"
	"go-service-template/internal/model"
	"go-service-template/internal/repository"
)

func newStore(t *testing.T) *repository.GormItemStore {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "repo.db"))
	if err != nil {
		t.Fatalf("database: %v", err)
	}
	return repository.NewItemStore(db)
}

func TestCreateGetListUpdateDelete(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	item := model.Item{Name: "first", Status: model.StatusActive}
	if err := store.Create(ctx, &item); err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.ID == 0 || item.CreatedAt.IsZero() {
		t.Fatalf("generated fields missing: %+v", item)
	}

	got, err := store.Get(ctx, item.ID)
	if err != nil || got.Name != "first" {
		t.Fatalf("get: %+v %v", got, err)
	}

	second := model.Item{Name: "second", Status: model.StatusArchived}
	if err := store.Create(ctx, &second); err != nil {
		t.Fatalf("create second: %v", err)
	}

	items, total, err := store.List(ctx, 1, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(items) != 1 {
		t.Fatalf("expected total 2 and one page row, got total=%d rows=%d", total, len(items))
	}
	if items[0].Name != "second" {
		t.Fatalf("expected newest first, got %q", items[0].Name)
	}

	got.Name = "renamed"
	if err := store.Update(ctx, &got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if refreshed, _ := store.Get(ctx, got.ID); refreshed.Name != "renamed" {
		t.Fatalf("update was not persisted: %+v", refreshed)
	}

	if err := store.Delete(ctx, got.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(ctx, got.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMissingRowsReturnErrNotFound(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	if _, err := store.Get(ctx, 404); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("get: %v", err)
	}
	if err := store.Delete(ctx, 404); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("delete: %v", err)
	}
	if err := store.Update(ctx, &model.Item{ID: 404, Name: "ghost"}); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("update: %v", err)
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatalf("database: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := database.Seed(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	var count int64
	if err := db.Model(&model.Item{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Fatalf("seed ran twice: count = %d", count)
	}
}
