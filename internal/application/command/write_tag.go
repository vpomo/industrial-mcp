package command

import (
	"context"
	"encoding/json"

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

func (h *WriteTagHandler) Handle(ctx context.Context, cmd WriteTagCommand) (*WriteTagResponse, error) {
	tag, err := entity.NewTag(cmd.TagName, cmd.Value)
	if err != nil {
		return nil, err
	}
	if err := h.repo.Save(ctx, tag); err != nil {
		return nil, err
	}
	if h.publisher != nil {
		payload, _ := json.Marshal(tag)
		h.publisher.Publish(ctx, "mcp/tag/written", payload)
	}
	return &WriteTagResponse{ID: tag.ID(), Success: true}, nil
}
