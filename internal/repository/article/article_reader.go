package article

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type ReaderRepository interface {
	// Save 有就更新，没有就创建，即 upsert
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type CachedReaderRepo struct {
}

func NewCachedReaderRepo() ReaderRepository {
	return &CachedReaderRepo{}
}

func (c *CachedReaderRepo) Save(ctx context.Context, art domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}
