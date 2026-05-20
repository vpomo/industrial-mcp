package driver

import (
	"context"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/entity"
)

type ModbusDriver struct {
	config map[string]string
}

func NewModbusDriver() *ModbusDriver {
	return &ModbusDriver{}
}

func (d *ModbusDriver) Type() entity.DataSourceType {
	return entity.DataSourceTypeModbus
}

func (d *ModbusDriver) Connect(ctx context.Context, config map[string]string) error {
	d.config = config
	return nil
}

func (d *ModbusDriver) Disconnect() {
	d.config = nil
}

func (d *ModbusDriver) ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error) {
	return nil, ErrNotImplemented
}

func (d *ModbusDriver) WriteTag(ctx context.Context, nodeID string, value interface{}) error {
	return ErrNotImplemented
}

func (d *ModbusDriver) Scan(ctx context.Context) ([]ScanResult, error) {
	return nil, ErrNotImplemented
}
