package query

import (
	"context"
	"testing"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
	"github.com/vpomo/industrial-mcp/internal/domain/service"
	infrarepo "github.com/vpomo/industrial-mcp/internal/infrastructure/repository"
)

func TestReadTagHandler(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	svc := service.NewTagService(repo)
	h := NewReadTagHandler(svc)

	tag, _ := entity.NewTag("temp_sensor", 23.5)
	repo.Save(context.Background(), tag)

	resp, err := h.Handle(context.Background(), ReadTagQuery{TagName: "temp_sensor"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != "temp_sensor" {
		t.Errorf("expected name 'temp_sensor', got %s", resp.Name)
	}
	if resp.Value != 23.5 {
		t.Errorf("expected value 23.5, got %v", resp.Value)
	}
	if resp.Quality != int(entity.QualityGood) {
		t.Errorf("expected quality QualityGood, got %d", resp.Quality)
	}
}

func TestReadTagHandlerNotFound(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	svc := service.NewTagService(repo)
	h := NewReadTagHandler(svc)

	_, err := h.Handle(context.Background(), ReadTagQuery{TagName: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent tag")
	}
}

func TestListTagsHandler(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	svc := service.NewTagService(repo)
	h := NewListTagsHandler(svc)

	tags := []*entity.Tag{
		{}, {},
	}
	tags[0], _ = entity.NewTag("sensor1", 10.0)
	tags[1], _ = entity.NewTag("sensor2", 20.0)
	for _, tag := range tags {
		repo.Save(context.Background(), tag)
	}

	resp, err := h.Handle(context.Background(), ListTagsQuery{Limit: 10})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
	if len(resp.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(resp.Tags))
	}
}

func TestListTagsHandlerEmpty(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	svc := service.NewTagService(repo)
	h := NewListTagsHandler(svc)

	resp, err := h.Handle(context.Background(), ListTagsQuery{Limit: 10})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}
