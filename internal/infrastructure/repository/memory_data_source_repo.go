package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/entity"
)

var ErrDataSourceNotFound = errors.New("data source not found")

type MemoryDataSourceRepository struct {
	mu          sync.RWMutex
	dataSources map[string]*entity.DataSource
}

func NewMemoryDataSourceRepository() *MemoryDataSourceRepository {
	return &MemoryDataSourceRepository{dataSources: make(map[string]*entity.DataSource)}
}

func (r *MemoryDataSourceRepository) GetByID(ctx context.Context, id string) (*entity.DataSource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, ok := r.dataSources[id]
	if !ok {
		return nil, ErrDataSourceNotFound
	}
	return ds, nil
}

func (r *MemoryDataSourceRepository) List(ctx context.Context) ([]*entity.DataSource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entity.DataSource, 0, len(r.dataSources))
	for _, ds := range r.dataSources {
		result = append(result, ds)
	}
	return result, nil
}

func (r *MemoryDataSourceRepository) Save(ctx context.Context, ds *entity.DataSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dataSources[ds.ID()] = ds
	return nil
}

func (r *MemoryDataSourceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.dataSources, id)
	return nil
}
