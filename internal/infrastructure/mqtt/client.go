package mqtt

import (
	"context"
	"reflect"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type TagSubscriber interface {
	Subscribe(ctx context.Context, topic string, handler func([]byte)) error
}

type PublishPublisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

type MQTTClient struct {
	client mqtt.Client
	prefix string
	mu     sync.RWMutex
}

func NewMQTTClient(brokerURL, clientID, prefix string) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &MQTTClient{client: client, prefix: prefix}, nil
}

func (m *MQTTClient) Publish(ctx context.Context, topic string, payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fullTopic := m.prefix + topic
	token := m.client.Publish(fullTopic, 0, false, payload)
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Subscribe(ctx context.Context, topic string, handler func([]byte)) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fullTopic := m.prefix + topic
	wrappedHandler := func(client mqtt.Client, msg mqtt.Message) {
		handler(msg.Payload())
	}
	token := m.client.Subscribe(fullTopic, 0, wrappedHandler)
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Disconnect() {
	if m.client != nil {
		m.client.Disconnect(250)
	}
}
