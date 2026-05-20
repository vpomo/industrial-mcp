package driver

import (
	"context"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/entity"
)

type MQTTDriver struct {
	config map[string]string
}

func NewMQTTDriver() *MQTTDriver {
	return &MQTTDriver{}
}

func (d *MQTTDriver) Type() entity.DataSourceType {
	return entity.DataSourceTypeMQTT
}

func (d *MQTTDriver) Connect(ctx context.Context, config map[string]string) error {
	d.config = config
	return nil
}

func (d *MQTTDriver) Disconnect() {
	d.config = nil
}

func (d *MQTTDriver) ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error) {
	return nil, ErrNotImplemented
}

func (d *MQTTDriver) WriteTag(ctx context.Context, nodeID string, value interface{}) error {
	return ErrNotImplemented
}

func (d *MQTTDriver) Scan(ctx context.Context) ([]ScanResult, error) {
	return nil, ErrNotImplemented
}

var ErrNotImplemented = &DriverError{Message: "not implemented"}
