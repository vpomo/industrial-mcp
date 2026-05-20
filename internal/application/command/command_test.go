package command

import (
	"context"
	"errors"
	"sync"
	"testing"

	infrarepo "github.com/vpomo/industrial-mcp/internal/infrastructure/repository"
)

type mockPublisher struct {
	mu        sync.Mutex
	published []struct {
		topic   string
		payload []byte
	}
	err error
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, struct {
		topic   string
		payload []byte
	}{topic, payload})
	return nil
}

func TestWriteTagHandler(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	publisher := &mockPublisher{}
	h := NewWriteTagHandler(repo, publisher)

	resp, err := h.Handle(context.Background(), WriteTagCommand{
		TagName: "temperature",
		Value:   25.5,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.ID == "" {
		t.Error("expected non-empty ID")
	}

	tag, err := repo.GetByName(context.Background(), "temperature")
	if err != nil {
		t.Fatalf("expected tag to be saved, got error %v", err)
	}
	if tag.Value() != 25.5 {
		t.Errorf("expected value 25.5, got %v", tag.Value())
	}

	publisher.mu.Lock()
	if len(publisher.published) != 1 {
		t.Errorf("expected 1 published message, got %d", len(publisher.published))
	}
	if publisher.published[0].topic != "mcp/tag/written" {
		t.Errorf("expected topic 'mcp/tag/written', got %s", publisher.published[0].topic)
	}
	publisher.mu.Unlock()
}

func TestWriteTagHandlerNilPublisher(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	h := NewWriteTagHandler(repo, nil)

	resp, err := h.Handle(context.Background(), WriteTagCommand{
		TagName: "sensor",
		Value:   100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestWriteTagHandlerEmptyName(t *testing.T) {
	repo := infrarepo.NewMemoryTagRepository()
	publisher := &mockPublisher{}
	h := NewWriteTagHandler(repo, publisher)

	_, err := h.Handle(context.Background(), WriteTagCommand{
		TagName: "",
		Value:   25.5,
	})
	if err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

type mockSubscriber struct {
	err error
}

func (m *mockSubscriber) Subscribe(ctx context.Context, topic string, handler func([]byte)) error {
	return m.err
}

func TestSubscribeTagHandler(t *testing.T) {
	subscriber := &mockSubscriber{}
	h := NewSubscribeTagHandler(subscriber)

	resp, err := h.Handle(context.Background(), SubscribeTagCommand{
		TagName: "my_sensor",
		QoS:     0,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.SubscriptionID == "" {
		t.Error("expected non-empty subscription ID")
	}
	if resp.Topic != "mcp/tag/my_sensor" {
		t.Errorf("expected topic 'mcp/tag/my_sensor', got %s", resp.Topic)
	}
}

func TestSubscribeTagHandlerError(t *testing.T) {
	subscriber := &mockSubscriber{err: errors.New("connection failed")}
	h := NewSubscribeTagHandler(subscriber)

	_, err := h.Handle(context.Background(), SubscribeTagCommand{
		TagName: "sensor",
		QoS:     0,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
