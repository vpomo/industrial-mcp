package query

import (
	"context"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/service"
)

type ListTagsQuery struct {
	Limit int `json:"limit"`
}

type ListTagsHandler struct {
	tagService *service.TagService
}

func NewListTagsHandler(ts *service.TagService) *ListTagsHandler {
	return &ListTagsHandler{tagService: ts}
}

type ListTagsResponse struct {
	Tags []TagInfo `json:"tags"`
	Count int      `json:"count"`
}

type TagInfo struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Quality   int         `json:"quality"`
}

func (h *ListTagsHandler) Handle(ctx context.Context, q ListTagsQuery) (*ListTagsResponse, error) {
	tags, err := h.tagService.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]TagInfo, 0, len(tags))
	for _, tag := range tags {
		result = append(result, TagInfo{
			ID:      tag.ID(),
			Name:    tag.Name(),
			Value:   tag.Value(),
			Quality: int(tag.Quality()),
		})
	}

	return &ListTagsResponse{
		Tags:  result,
		Count: len(result),
	}, nil
}