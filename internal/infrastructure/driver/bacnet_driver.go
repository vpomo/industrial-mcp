package driver

import (
	"context"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
)

type BACnetDriver struct {
	config map[string]string
}

func NewBACnetDriver() *BACnetDriver {
	return &BACnetDriver{}
}

func (d *BACnetDriver) Type() entity.DataSourceType {
	return entity.DataSourceTypeBACnet
}

func (d *BACnetDriver) Connect(ctx context.Context, config map[string]string) error {
	d.config = config
	return nil
}

func (d *BACnetDriver) Disconnect() {
	d.config = nil
}

func (d *BACnetDriver) ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error) {
	return nil, ErrNotImplemented
}

func (d *BACnetDriver) WriteTag(ctx context.Context, nodeID string, value interface{}) error {
	return ErrNotImplemented
}

func (d *BACnetDriver) Scan(ctx context.Context) ([]ScanResult, error) {
	return nil, ErrNotImplemented
}
