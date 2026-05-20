package repository

import (
	"context"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
)

type TagReader interface {
	GetByID(ctx context.Context, id string) (*entity.Tag, error)
	GetByName(ctx context.Context, name string) (*entity.Tag, error)
	List(ctx context.Context) ([]*entity.Tag, error)
}

type TagWriter interface {
	Save(ctx context.Context, tag *entity.Tag) error
	Delete(ctx context.Context, id string) error
}

type TagRepository interface {
	TagReader
	TagWriter
}
