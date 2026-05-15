package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Quality int

const (
	QualityGood       Quality = 0
	QualityUncertain Quality = 1
	QualityBad       Quality = 2
)

type Tag struct {
	id        string
	name      string
	value     interface{}
	timestamp time.Time
	quality   Quality
}

func NewTag(name string, value interface{}) (*Tag, error) {
	if name == "" {
		return nil, errors.New("tag name is required")
	}
	return &Tag{
		id:        uuid.New().String(),
		name:      name,
		value:     value,
		timestamp: time.Now(),
		quality:   QualityGood,
	}, nil
}

func (t *Tag) ID() string          { return t.id }
func (t *Tag) Name() string        { return t.name }
func (t *Tag) Value() interface{}  { return t.value }
func (t *Tag) Timestamp() time.Time { return t.timestamp }
func (t *Tag) Quality() Quality    { return t.quality }

func (t *Tag) UpdateValue(newValue interface{}) {
	t.value = newValue
	t.timestamp = time.Now()
}

func (t *Tag) SetQuality(q Quality) {
	t.quality = q
}