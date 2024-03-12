package service

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/repository"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type articleSvc struct {
	repo repository.ArticleRepository
}

func NewArticleSvc(repo repository.ArticleRepository) ArticleService {
	return &articleSvc{
		repo: repo,
	}
}

func (a *articleSvc) Save(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Create(ctx, art)
}
