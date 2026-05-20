package service

import (
	"context"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
	"github.com/vpomo/industrial-mcp/internal/domain/repository"
	"github.com/vpomo/industrial-mcp/internal/infrastructure/driver"
)

type DataSourceService struct {
	repo      repository.DataSourceRepository
	driverMgr *driver.DriverManager
	connected map[string]driver.DataSourceDriver
}

func NewDataSourceService(repo repository.DataSourceRepository, driverMgr *driver.DriverManager) *DataSourceService {
	return &DataSourceService{
		repo:      repo,
		driverMgr: driverMgr,
		connected: make(map[string]driver.DataSourceDriver),
	}
}

func (s *DataSourceService) Create(ctx context.Context, name string, dsType entity.DataSourceType, config map[string]string) (*entity.DataSource, error) {
	ds, err := entity.NewDataSource(name, dsType, config)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DataSourceService) GetByID(ctx context.Context, id string) (*entity.DataSource, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DataSourceService) List(ctx context.Context) ([]*entity.DataSource, error) {
	return s.repo.List(ctx)
}

func (s *DataSourceService) Delete(ctx context.Context, id string) error {
	s.Disconnect(ctx, id)
	return s.repo.Delete(ctx, id)
}

func (s *DataSourceService) Connect(ctx context.Context, id string) error {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	drv, ok := s.driverMgr.GetDriver(ds.Type())
	if !ok {
		return ErrDriverNotFound
	}

	if err := drv.Connect(ctx, ds.Config()); err != nil {
		return err
	}

	s.connected[id] = drv
	return nil
}

func (s *DataSourceService) Disconnect(ctx context.Context, id string) error {
	if drv, ok := s.connected[id]; ok {
		drv.Disconnect()
		delete(s.connected, id)
	}
	return nil
}

func (s *DataSourceService) Scan(ctx context.Context, id string) ([]driver.ScanResult, error) {
	drv, ok := s.connected[id]
	if !ok {
		return nil, ErrNotConnected
	}
	return drv.Scan(ctx)
}

func (s *DataSourceService) ReadTag(ctx context.Context, id, nodeID string) (*entity.Tag, error) {
	drv, ok := s.connected[id]
	if !ok {
		return nil, ErrNotConnected
	}
	return drv.ReadTag(ctx, nodeID)
}

func (s *DataSourceService) WriteTag(ctx context.Context, id, nodeID string, value interface{}) error {
	drv, ok := s.connected[id]
	if !ok {
		return ErrNotConnected
	}
	return drv.WriteTag(ctx, nodeID, value)
}
