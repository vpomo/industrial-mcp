package repository

import (
	"context"
	"testing"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/entity"
)

func TestMemoryTagRepositorySaveAndGet(t *testing.T) {
	repo := NewMemoryTagRepository()
	ctx := context.Background()

	tag, _ := entity.NewTag("test_tag", 42.0)
	err := repo.Save(ctx, tag)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found, err := repo.GetByID(ctx, tag.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Name() != "test_tag" {
		t.Errorf("expected 'test_tag', got %s", found.Name())
	}
}

func TestMemoryTagRepositoryGetByName(t *testing.T) {
	repo := NewMemoryTagRepository()
	ctx := context.Background()

	tag, _ := entity.NewTag("my_sensor", 99.9)
	repo.Save(ctx, tag)

	found, err := repo.GetByName(ctx, "my_sensor")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Value() != 99.9 {
		t.Errorf("expected 99.9, got %v", found.Value())
	}
}

func TestMemoryTagRepositoryGetByNameNotFound(t *testing.T) {
	repo := NewMemoryTagRepository()
	ctx := context.Background()

	_, err := repo.GetByName(ctx, "nonexistent")
	if err != ErrTagNotFound {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestMemoryTagRepositoryList(t *testing.T) {
	repo := NewMemoryTagRepository()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		tag, _ := entity.NewTag("sensor", float64(i)*10.0)
		repo.Save(ctx, tag)
	}

	tags, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

func TestMemoryTagRepositoryDelete(t *testing.T) {
	repo := NewMemoryTagRepository()
	ctx := context.Background()

	tag, _ := entity.NewTag("to_delete", 10.0)
	repo.Save(ctx, tag)

	err := repo.Delete(ctx, tag.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.GetByID(ctx, tag.ID())
	if err != ErrTagNotFound {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestMemoryTagRepositoryConcurrent(t *testing.T) {
	repo := NewMemoryTagRepository()
	ctx := context.Background()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			tag, _ := entity.NewTag("concurrent", float64(idx))
			repo.Save(ctx, tag)
			repo.GetByName(ctx, "concurrent")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
