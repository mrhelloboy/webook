package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
}

type gormArticleDAO struct {
	db *gorm.DB
}

func NewGormArticleDAO(db *gorm.DB) ArticleDAO {
	return &gormArticleDAO{db: db}
}

func (g *gormArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := g.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// Article 制作库
type Article struct {
	Id       int64  `gorm:"primary_key,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index"`
	Ctime    int64
	Utime    int64
}
