package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/entity"
)

var ErrExposedTagNotFound = errors.New("exposed tag not found")

type MemoryExposedTagRepository struct {
	mu    sync.RWMutex
	tags  map[string]*entity.ExposedTag
	index map[string][]*entity.ExposedTag
}

func NewMemoryExposedTagRepository() *MemoryExposedTagRepository {
	return &MemoryExposedTagRepository{
		tags:  make(map[string]*entity.ExposedTag),
		index: make(map[string][]*entity.ExposedTag),
	}
}

func (r *MemoryExposedTagRepository) GetByID(ctx context.Context, id string) (*entity.ExposedTag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tag, ok := r.tags[id]
	if !ok {
		return nil, ErrExposedTagNotFound
	}
	return tag, nil
}

func (r *MemoryExposedTagRepository) GetByDataSourceID(ctx context.Context, dataSourceID string) ([]*entity.ExposedTag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.index[dataSourceID], nil
}

func (r *MemoryExposedTagRepository) List(ctx context.Context) ([]*entity.ExposedTag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entity.ExposedTag, 0, len(r.tags))
	for _, tag := range r.tags {
		result = append(result, tag)
	}
	return result, nil
}

func (r *MemoryExposedTagRepository) Save(ctx context.Context, tag *entity.ExposedTag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags[tag.ID()] = tag
	r.index[tag.DataSourceID()] = append(r.index[tag.DataSourceID()], tag)
	return nil
}

func (r *MemoryExposedTagRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tag, ok := r.tags[id]
	if !ok {
		return nil
	}
	delete(r.tags, id)
	newIndex := make([]*entity.ExposedTag, 0)
	for _, t := range r.index[tag.DataSourceID()] {
		if t.ID() != id {
			newIndex = append(newIndex, t)
		}
	}
	r.index[tag.DataSourceID()] = newIndex
	return nil
}
