package repository

import (
	"context"

	"github.com/imatic/mcp_mqtt_opcua/internal/domain/entity"
)

type DataSourceReader interface {
	GetByID(ctx context.Context, id string) (*entity.DataSource, error)
	List(ctx context.Context) ([]*entity.DataSource, error)
}

type DataSourceWriter interface {
	Save(ctx context.Context, ds *entity.DataSource) error
	Delete(ctx context.Context, id string) error
}

type DataSourceRepository interface {
	DataSourceReader
	DataSourceWriter
}
