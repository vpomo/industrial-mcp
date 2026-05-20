package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ExposedTag struct {
	id           string
	name         string
	dataSourceID string
	nodeID       string
	readOnly     bool
	dataType     string
	createdAt    time.Time
}

func NewExposedTag(name, dataSourceID, nodeID, dataType string, readOnly bool) (*ExposedTag, error) {
	if name == "" {
		return nil, errors.New("tag name is required")
	}
	if dataSourceID == "" {
		return nil, errors.New("data source ID is required")
	}
	if nodeID == "" {
		return nil, errors.New("node ID is required")
	}
	return &ExposedTag{
		id:           uuid.New().String(),
		name:         name,
		dataSourceID: dataSourceID,
		nodeID:       nodeID,
		readOnly:     readOnly,
		dataType:     dataType,
		createdAt:    time.Now(),
	}, nil
}

func (t *ExposedTag) ID() string           { return t.id }
func (t *ExposedTag) Name() string         { return t.name }
func (t *ExposedTag) DataSourceID() string { return t.dataSourceID }
func (t *ExposedTag) NodeID() string       { return t.nodeID }
func (t *ExposedTag) ReadOnly() bool       { return t.readOnly }
func (t *ExposedTag) DataType() string     { return t.dataType }
func (t *ExposedTag) CreatedAt() time.Time { return t.createdAt }

func (t *ExposedTag) SetName(name string)         { t.name = name }
func (t *ExposedTag) SetNodeID(nodeID string)     { t.nodeID = nodeID }
func (t *ExposedTag) SetReadOnly(readOnly bool)   { t.readOnly = readOnly }
func (t *ExposedTag) SetDataType(dataType string) { t.dataType = dataType }
