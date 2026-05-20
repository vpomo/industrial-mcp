package service

import (
	"context"
	"errors"

	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/entity"
	"github.com/vpomo/mcp_mqtt_opcua/internal/domain/repository"
)

var (
	ErrDriverNotFound       = errors.New("driver not found for data source type")
	ErrNotConnected         = errors.New("data source not connected")
	ErrDataSourceIDRequired = errors.New("data source ID is required")
)

type ExposedTagService struct {
	repo      repository.ExposedTagRepository
	dsService *DataSourceService
}

func NewExposedTagService(repo repository.ExposedTagRepository, dsService *DataSourceService) *ExposedTagService {
	return &ExposedTagService{
		repo:      repo,
		dsService: dsService,
	}
}

func (s *ExposedTagService) Create(ctx context.Context, name, dataSourceID, nodeID, dataType string, readOnly bool) (*entity.ExposedTag, error) {
	if dataSourceID == "" {
		return nil, ErrDataSourceIDRequired
	}
	tag, err := entity.NewExposedTag(name, dataSourceID, nodeID, dataType, readOnly)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *ExposedTagService) GetByID(ctx context.Context, id string) (*entity.ExposedTag, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ExposedTagService) List(ctx context.Context) ([]*entity.ExposedTag, error) {
	return s.repo.List(ctx)
}

func (s *ExposedTagService) ListByDataSource(ctx context.Context, dataSourceID string) ([]*entity.ExposedTag, error) {
	return s.repo.GetByDataSourceID(ctx, dataSourceID)
}

func (s *ExposedTagService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ExposedTagService) ReadValue(ctx context.Context, id string) (*entity.Tag, error) {
	exposedTag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.dsService.ReadTag(ctx, exposedTag.DataSourceID(), exposedTag.NodeID())
}

func (s *ExposedTagService) WriteValue(ctx context.Context, id string, value interface{}) error {
	exposedTag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if exposedTag.ReadOnly() {
		return errors.New("tag is read-only")
	}
	return s.dsService.WriteTag(ctx, exposedTag.DataSourceID(), exposedTag.NodeID(), value)
}
