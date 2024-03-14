package article

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art PublishArticle) error
}

type gormReaderDAO struct {
	db *gorm.DB
}

func NewGormReaderDAO(db *gorm.DB) ReaderDAO {
	return &gormReaderDAO{db: db}
}

// Upsert 插入或更新文章
func (g *gormReaderDAO) Upsert(ctx context.Context, art PublishArticle) error {
	err := g.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
		}),
	}).Create(&art).Error
	return err
}

// PublishArticle 线上库，表结构跟制作库一致
type PublishArticle struct {
	Article
}
