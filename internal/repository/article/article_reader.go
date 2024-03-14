package article

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/internal/domain"
	daoArt "github.com/mrhelloboy/wehook/internal/repository/dao/article"
)

type ReaderRepository interface {
	// Save 有就更新，没有就创建，即 upsert
	Save(ctx context.Context, art domain.Article) error
}

type CachedReaderRepo struct {
	dao daoArt.ReaderDAO
}

func NewCachedReaderRepo(dao daoArt.ReaderDAO) ReaderRepository {
	return &CachedReaderRepo{dao: dao}
}

func (c *CachedReaderRepo) Save(ctx context.Context, art domain.Article) error {
	now := time.Now().UnixMilli()

	return c.dao.Upsert(ctx, daoArt.PublishArticle{Article: daoArt.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Ctime:    now,
		Utime:    now,
	}})
}
