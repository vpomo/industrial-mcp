package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type DataSourceType string

const (
	DataSourceTypeOPCUA  DataSourceType = "opcua"
	DataSourceTypeMQTT   DataSourceType = "mqtt"
	DataSourceTypeModbus DataSourceType = "modbus"
	DataSourceTypeBACnet DataSourceType = "bacnet"
)

func (t DataSourceType) IsValid() bool {
	switch t {
	case DataSourceTypeOPCUA, DataSourceTypeMQTT, DataSourceTypeModbus, DataSourceTypeBACnet:
		return true
	}
	return false
}

type DataSource struct {
	id        string
	name      string
	dsType    DataSourceType
	config    map[string]string
	enabled   bool
	createdAt time.Time
	updatedAt time.Time
}

func NewDataSource(name string, dsType DataSourceType, config map[string]string) (*DataSource, error) {
	if name == "" {
		return nil, errors.New("data source name is required")
	}
	if !dsType.IsValid() {
		return nil, errors.New("invalid data source type")
	}
	if config == nil {
		config = make(map[string]string)
	}
	return &DataSource{
		id:        uuid.New().String(),
		name:      name,
		dsType:    dsType,
		config:    config,
		enabled:   true,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

func (ds *DataSource) ID() string                { return ds.id }
func (ds *DataSource) Name() string              { return ds.name }
func (ds *DataSource) Type() DataSourceType      { return ds.dsType }
func (ds *DataSource) Config() map[string]string { return ds.config }
func (ds *DataSource) Enabled() bool             { return ds.enabled }
func (ds *DataSource) CreatedAt() time.Time      { return ds.createdAt }
func (ds *DataSource) UpdatedAt() time.Time      { return ds.updatedAt }

func (ds *DataSource) SetName(name string) { ds.name = name; ds.updatedAt = time.Now() }
func (ds *DataSource) SetConfig(config map[string]string) {
	ds.config = config
	ds.updatedAt = time.Now()
}
func (ds *DataSource) SetEnabled(enabled bool) { ds.enabled = enabled; ds.updatedAt = time.Now() }
