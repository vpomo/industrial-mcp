package query

import (
	"context"
	"time"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/service"
)

type ReadTagQuery struct {
	TagName string `json:"tag_name"`
}

type ReadTagHandler struct {
	tagService *service.TagService
}

func NewReadTagHandler(ts *service.TagService) *ReadTagHandler {
	return &ReadTagHandler{tagService: ts}
}

type ReadTagResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Timestamp string      `json:"timestamp"`
	Quality   int         `json:"quality"`
}

func (h *ReadTagHandler) Handle(ctx context.Context, q ReadTagQuery) (*ReadTagResponse, error) {
	tag, err := h.tagService.GetTag(ctx, q.TagName)
	if err != nil {
		return nil, err
	}
	return &ReadTagResponse{
		ID:        tag.ID(),
		Name:      tag.Name(),
		Value:     tag.Value(),
		Timestamp: tag.Timestamp().Format(time.RFC3339),
		Quality:   int(tag.Quality()),
	}, nil
}
