package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
	"github.com/vpomo/industrial-mcp/internal/domain/repository"
	"github.com/vpomo/industrial-mcp/internal/infrastructure/mqtt"
)

type WriteTagCommand struct {
	TagName string      `json:"tag_name"`
	Value   interface{} `json:"value"`
}

type WriteTagHandler struct {
	repo      repository.TagRepository
	publisher mqtt.PublishPublisher
}

func NewWriteTagHandler(repo repository.TagRepository, pub mqtt.PublishPublisher) *WriteTagHandler {
	return &WriteTagHandler{repo: repo, publisher: pub}
}

type WriteTagResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}

type tagMQTTPayload struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Timestamp string      `json:"timestamp"`
	Quality   int         `json:"quality"`
}

func (h *WriteTagHandler) Handle(ctx context.Context, cmd WriteTagCommand) (*WriteTagResponse, error) {
	tag, err := entity.NewTag(cmd.TagName, cmd.Value)
	if err != nil {
		return nil, err
	}
	if err := h.repo.Save(ctx, tag); err != nil {
		return nil, err
	}
	if mqtt.IsNil(h.publisher) {
		return &WriteTagResponse{ID: tag.ID(), Success: true}, nil
	}

	payload, err := json.Marshal(tagMQTTPayload{
		ID:        tag.ID(),
		Name:      tag.Name(),
		Value:     tag.Value(),
		Timestamp: tag.Timestamp().Format(time.RFC3339),
		Quality:   int(tag.Quality()),
	})
	if err != nil {
		return nil, err
	}

	tagTopic := "tag/" + tag.Name()
	if err := h.publisher.Publish(ctx, tagTopic, payload); err != nil {
		return nil, fmt.Errorf("mqtt publish tag: %w", err)
	}
	if err := h.publisher.Publish(ctx, "tag/written", payload); err != nil {
		return nil, fmt.Errorf("mqtt publish written: %w", err)
	}

	return &WriteTagResponse{ID: tag.ID(), Success: true}, nil
}
