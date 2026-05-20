package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/vpomo/mcp_mqtt_opcua/internal/infrastructure/mqtt"
)

type SubscribeTagCommand struct {
	TagName string `json:"tag_name"`
	QoS     int    `json:"qos"`
}

type SubscribeTagHandler struct {
	subscriber mqtt.TagSubscriber
}

func NewSubscribeTagHandler(sub mqtt.TagSubscriber) *SubscribeTagHandler {
	return &SubscribeTagHandler{subscriber: sub}
}

type SubscribeTagResponse struct {
	SubscriptionID string `json:"subscription_id"`
	Topic          string `json:"topic"`
}

func (h *SubscribeTagHandler) Handle(ctx context.Context, cmd SubscribeTagCommand) (*SubscribeTagResponse, error) {
	subID := uuid.New().String()
	topic := "mcp/tag/" + cmd.TagName
	err := h.subscriber.Subscribe(ctx, topic, func(data []byte) {
	})
	if err != nil {
		return nil, err
	}
	return &SubscribeTagResponse{
		SubscriptionID: subID,
		Topic:          topic,
	}, nil
}
