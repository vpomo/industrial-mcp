package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/entity"
)

var ErrTagNotFound = errors.New("tag not found")

type MemoryTagRepository struct {
	mu   sync.RWMutex
	tags map[string]*entity.Tag
}

func NewMemoryTagRepository() *MemoryTagRepository {
	return &MemoryTagRepository{tags: make(map[string]*entity.Tag)}
}

func (r *MemoryTagRepository) GetByID(ctx context.Context, id string) (*entity.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tag, ok := r.tags[id]
	if !ok {
		return nil, ErrTagNotFound
	}
	return tag, nil
}

func (r *MemoryTagRepository) GetByName(ctx context.Context, name string) (*entity.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tag := range r.tags {
		if tag.Name() == name {
			return tag, nil
		}
	}
	return nil, ErrTagNotFound
}

func (r *MemoryTagRepository) List(ctx context.Context) ([]*entity.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entity.Tag, 0, len(r.tags))
	for _, tag := range r.tags {
		result = append(result, tag)
	}
	return result, nil
}

func (r *MemoryTagRepository) Save(ctx context.Context, tag *entity.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tags[tag.ID()] = tag
	return nil
}

func (r *MemoryTagRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tags, id)
	return nil
}