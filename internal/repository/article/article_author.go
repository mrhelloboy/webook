package article

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/repository/dao"

	"github.com/mrhelloboy/wehook/internal/domain"
)

// AuthorRepository 制作库接口
type AuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

type cachedAuthorRepo struct {
	dao dao.ArticleDAO
}

func NewCachedAuthorRepo(dao dao.ArticleDAO) AuthorRepository {
	return &cachedAuthorRepo{
		dao: dao,
	}
}

func (c *cachedAuthorRepo) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (c *cachedAuthorRepo) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}
