package mcp

import (
	"context"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/service"
)

type TagServiceWrapper struct {
	tagService *service.TagService
	mqttClient MQTTClientWrapper
}

type MQTTClientWrapper interface {
	Publish(ctx context.Context, topic string, payload []byte) error
	Subscribe(ctx context.Context, topic string, handler func([]byte)) error
}

func NewTagServiceWrapper(ts *service.TagService, mqtt MQTTClientWrapper) *TagServiceWrapper {
	return &TagServiceWrapper{
		tagService: ts,
		mqttClient: mqtt,
	}
}

func (w *TagServiceWrapper) GetTag(ctx context.Context, name string) (*TagResponse, error) {
	tag, err := w.tagService.GetTag(ctx, name)
	if err != nil {
		return nil, err
	}
	return &TagResponse{
		ID:    tag.ID(),
		Name:  tag.Name(),
		Value: tag.Value(),
	}, nil
}

type TagResponse struct {
	ID    string
	Name  string
	Value interface{}
}