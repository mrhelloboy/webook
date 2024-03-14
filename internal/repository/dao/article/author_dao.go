package article

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthorDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishedArticle) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
}

type gormAuthorDAO struct {
	db *gorm.DB
}

func (g *gormAuthorDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id= ? and author = ?", id, author).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 0 {
			// 要么 ID 是错的，要么作者不对
			// 如果是作者不对，就需要留意是否有人在搞事情。
			// todo: 用 prometheus 打点，只要频繁出现，就需要告警，然后人为介入排查
			return fmt.Errorf("更新帖子状态失败，可能是创作者非法 id %d, author_id %d", id, author)
		}
		// 同步线上库的状态
		return tx.Model(&PublishedArticle{}).Where("id = ?", id).Updates(map[string]any{
			"status": status,
			"utime":  now,
		}).Error
	})
}

func (g *gormAuthorDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := g.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (g *gormAuthorDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 确保只有作者才可以修改
	res := g.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   art.Utime,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
	}
	return nil
}

func (g *gormAuthorDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作表，再操作线上表
	id := art.Id
	// 采用 GORM 的事务处理同步
	// GORM 进行了事务生命周期的管理，Begin，Rollback，Commit 都不需要我们操心
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGormArticleDAO(tx)
		// 先在制作库上进行更新或者新建
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 更新线上库
		return txDAO.Upsert(ctx, PublishedArticle{Article: art})
	})
	return id, err
}

// Upsert 对线上库进行更新或者创建操作
// Upsert: 对应数据库的 INSERT or UPDATE 操作
func (g *gormAuthorDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	// SQL: INSERT xxx ON DUPLICATE KEY UPDATE xxx
	err := g.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// INSERT aaa ON CONFLICT(bbb) DO NOTHING
		// INSERT aaa ON CONFLICT(bbb) DO UPDATE SET ccc WHERE ddd

		// 冲突字段
		//Columns: []clause.Column{{Name: "id"}},
		// 数据冲突时，什么也不做
		//DoNothing: true,
		// 数据冲突时，并且符合 WHERE 条件时就会执行 DO UPDATE
		//Where: clause.Where{
		//}
		// MySQL 只需使用 DoUpdates 字段
		DoUpdates: clause.Assignments(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&art).Error
	// 一条 SQL 语句都不需要开启事务
	// auto commit: 自动提交
	return err
}

func NewGormArticleDAO(db *gorm.DB) AuthorDAO {
	return &gormAuthorDAO{db: db}
}

// Article 制作库
type Article struct {
	Id       int64  `gorm:"primary_key,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index"`
	Status   uint8
	Ctime    int64
	Utime    int64
}
