package repository

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/repository/dao"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
}

type cachedArticleRepo struct {
	dao dao.ArticleDAO
}

func NewCachedArticleRepo(dao dao.ArticleDAO) ArticleRepository {
	return &cachedArticleRepo{
		dao: dao,
	}
}

func (c *cachedArticleRepo) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}
