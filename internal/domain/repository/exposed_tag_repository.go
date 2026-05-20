package repository

import (
	"context"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
)

type ExposedTagReader interface {
	GetByID(ctx context.Context, id string) (*entity.ExposedTag, error)
	GetByDataSourceID(ctx context.Context, dataSourceID string) ([]*entity.ExposedTag, error)
	List(ctx context.Context) ([]*entity.ExposedTag, error)
}

type ExposedTagWriter interface {
	Save(ctx context.Context, tag *entity.ExposedTag) error
	Delete(ctx context.Context, id string) error
}

type ExposedTagRepository interface {
	ExposedTagReader
	ExposedTagWriter
}
