package mqtt

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var ErrBrokerURLRequired = errors.New("mqtt broker url is required")

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
	qos    byte
	retain bool
	mu     sync.RWMutex
}

func NewClient(cfg Config) (*MQTTClient, error) {
	if cfg.BrokerURL == "" {
		return nil, ErrBrokerURLRequired
	}

	clientID := cfg.ClientID
	if clientID == "" {
		clientID = "mcp_server"
	}

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	qos := cfg.QoS
	if qos > 2 {
		qos = 0
	}

	return &MQTTClient{
		client: client,
		prefix: cfg.TopicPrefix,
		qos:    qos,
		retain: cfg.Retain,
	}, nil
}

func NewMQTTClient(brokerURL, clientID, prefix string) (*MQTTClient, error) {
	return NewClient(Config{
		BrokerURL:   brokerURL,
		ClientID:    clientID,
		TopicPrefix: prefix,
	})
}

func (m *MQTTClient) Publish(ctx context.Context, topic string, payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fullTopic := m.prefix + topic
	token := m.client.Publish(fullTopic, m.qos, m.retain, payload)
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
	token := m.client.Subscribe(fullTopic, m.qos, wrappedHandler)
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Disconnect() {
	if m.client != nil {
		m.client.Disconnect(250)
	}
}
