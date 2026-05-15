package service

import (
	"context"
	"testing"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/entity"
	"github.com/imatic/mcp_mqtt_opcua/internal/infrastructure/repository"
)

func TestTagServiceGetTag(t *testing.T) {
	repo := repository.NewMemoryTagRepository()
	svc := NewTagService(repo)

	tag, _ := entity.NewTag("temp_sensor", 23.5)
	repo.Save(context.Background(), tag)

	found, err := svc.GetTag(context.Background(), "temp_sensor")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Name() != "temp_sensor" {
		t.Errorf("expected name 'temp_sensor', got %s", found.Name())
	}
	if found.Value() != 23.5 {
		t.Errorf("expected value 23.5, got %v", found.Value())
	}
}

func TestTagServiceGetTagNotFound(t *testing.T) {
	repo := repository.NewMemoryTagRepository()
	svc := NewTagService(repo)

	_, err := svc.GetTag(context.Background(), "nonexistent")
	if err != ErrTagNotFound {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestTagServiceUpdateTag(t *testing.T) {
	repo := repository.NewMemoryTagRepository()
	svc := NewTagService(repo)

	tag, _ := entity.NewTag("pressure", 100.0)
	repo.Save(context.Background(), tag)

	err := svc.UpdateTag(context.Background(), "pressure", 105.0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := svc.GetTag(context.Background(), "pressure")
	if updated.Value() != 105.0 {
		t.Errorf("expected value 105.0, got %v", updated.Value())
	}
}

func TestTagServiceUpdateTagNotFound(t *testing.T) {
	repo := repository.NewMemoryTagRepository()
	svc := NewTagService(repo)

	err := svc.UpdateTag(context.Background(), "nonexistent", 50.0)
	if err != ErrTagNotFound {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestTagServiceListTags(t *testing.T) {
	repo := repository.NewMemoryTagRepository()
	svc := NewTagService(repo)

	tag1, _ := entity.NewTag("sensor1", 10.0)
	tag2, _ := entity.NewTag("sensor2", 20.0)
	repo.Save(context.Background(), tag1)
	repo.Save(context.Background(), tag2)

	tags, err := svc.ListTags(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestTagServiceDeleteTag(t *testing.T) {
	repo := repository.NewMemoryTagRepository()
	svc := NewTagService(repo)

	tag, _ := entity.NewTag("to_delete", 50.0)
	repo.Save(context.Background(), tag)

	err := svc.DeleteTag(context.Background(), tag.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.GetTag(context.Background(), "to_delete")
	if err != ErrTagNotFound {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}