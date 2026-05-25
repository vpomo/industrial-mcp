package mqtt

import (
	"context"
	"sync"
)

type LazyClient struct {
	cfg    Config
	client *MQTTClient
	mu     sync.Mutex
}

func NewLazyClient(cfg Config) *LazyClient {
	return &LazyClient{cfg: cfg}
}

func (l *LazyClient) ensure() (*MQTTClient, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.client != nil {
		return l.client, nil
	}

	client, err := NewClient(l.cfg)
	if err != nil {
		return nil, err
	}
	l.client = client
	return l.client, nil
}

func (l *LazyClient) Publish(ctx context.Context, topic string, payload []byte) error {
	client, err := l.ensure()
	if err != nil {
		return err
	}
	return client.Publish(ctx, topic, payload)
}

func (l *LazyClient) Subscribe(ctx context.Context, topic string, handler func([]byte)) error {
	client, err := l.ensure()
	if err != nil {
		return err
	}
	return client.Subscribe(ctx, topic, handler)
}

func (l *LazyClient) Disconnect() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.client != nil {
		l.client.Disconnect()
		l.client = nil
	}
}
