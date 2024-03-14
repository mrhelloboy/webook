package article

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/domain"
	daoArt "github.com/mrhelloboy/wehook/internal/repository/dao/article"
)

// AuthorRepository 制作库接口
type AuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
}

type cachedAuthorRepo struct {
	dao daoArt.AuthorDAO
}

func NewCachedAuthorRepo(dao daoArt.AuthorDAO) AuthorRepository {
	return &cachedAuthorRepo{
		dao: dao,
	}
}

func (c *cachedAuthorRepo) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, daoArt.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (c *cachedAuthorRepo) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, daoArt.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (c *cachedAuthorRepo) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

func (c *cachedAuthorRepo) toEntity(article domain.Article) daoArt.Article {
	return daoArt.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	}
}
