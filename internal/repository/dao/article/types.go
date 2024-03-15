package article

import "context"

type AuthorDAO interface {
	GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	// upsert(ctx context.Context, art PublishedArticle) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
}

type ReaderDAO interface {
	Upsert(ctx context.Context, art PublishedArticle) error
}
