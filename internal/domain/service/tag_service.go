package service

import (
	"context"
	"errors"

	"github.com/vpomo/industrial-mcp/internal/domain/entity"
	"github.com/vpomo/industrial-mcp/internal/domain/repository"
)

var ErrTagNotFound = errors.New("tag not found")

type TagService struct {
	repo repository.TagRepository
}

func NewTagService(repo repository.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) GetTag(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, ErrTagNotFound
	}
	return tag, nil
}

func (s *TagService) UpdateTag(ctx context.Context, name string, value interface{}) error {
	tag, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return ErrTagNotFound
	}
	tag.UpdateValue(value)
	return s.repo.Save(ctx, tag)
}

func (s *TagService) ListTags(ctx context.Context) ([]*entity.Tag, error) {
	return s.repo.List(ctx)
}

func (s *TagService) SaveTag(ctx context.Context, tag *entity.Tag) error {
	return s.repo.Save(ctx, tag)
}

func (s *TagService) DeleteTag(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
