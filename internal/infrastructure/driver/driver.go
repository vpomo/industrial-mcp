package driver

import (
	"context"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/entity"
)

type ScanResult struct {
	NodeID   string
	Name     string
	DataType string
}

type DataSourceDriver interface {
	Type() entity.DataSourceType
	Connect(ctx context.Context, config map[string]string) error
	Disconnect()
	ReadTag(ctx context.Context, nodeID string) (*entity.Tag, error)
	WriteTag(ctx context.Context, nodeID string, value interface{}) error
	Scan(ctx context.Context) ([]ScanResult, error)
}

type DriverManager struct {
	drivers map[entity.DataSourceType]DataSourceDriver
}

func NewDriverManager() *DriverManager {
	return &DriverManager{drivers: make(map[entity.DataSourceType]DataSourceDriver)}
}

func (m *DriverManager) Register(driver DataSourceDriver) {
	m.drivers[driver.Type()] = driver
}

func (m *DriverManager) GetDriver(t entity.DataSourceType) (DataSourceDriver, bool) {
	driver, ok := m.drivers[t]
	return driver, ok
}

func (m *DriverManager) Unregister(t entity.DataSourceType) {
	delete(m.drivers, t)
}
