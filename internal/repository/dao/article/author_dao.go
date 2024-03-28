package article

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormAuthorDAO struct {
	db *gorm.DB
}

func (g *gormAuthorDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var res []Article
	err := g.db.WithContext(ctx).
		Where("utime < ?", start.UnixMilli()).
		Order("utime DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

// GetByAuthor 获取作者的文章列表 - 分页功能
func (g *gormAuthorDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	var arts []Article
	// todo: 数据库查询优化：author_id, utime 作为联合索引
	err := g.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (g *gormAuthorDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := g.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}

func (g *gormAuthorDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pubArt PublishedArticle
	err := g.db.WithContext(ctx).Where("id = ?", id).First(&pubArt).Error
	return pubArt, err
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
		// 为什么要创建一个新的 DAO 对象？
		// 事务内部会创建一个新的基于当前连接的 *gorm.DB 实例 tx，这个实例只在当前事务范围内有效，
		// 从而避免了直接在线程间共享 *gorm.DB 实例的问题。
		// 在事务内部都要通过 tx 来操作数据库，重新创建一个 DAO 对象可以避免直接操作 g.db，
		// 从而避免在事务内部共享 g.db 实例带来的问题。
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
		// todo：
		now := time.Now().UnixMilli()
		publishArt := PublishedArticle{Article: art}
		publishArt.Ctime = now
		publishArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
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
		}).Create(&publishArt).Error
		return err
	})
	return id, err
}

func NewGormArticleDAO(db *gorm.DB) AuthorDAO {
	return &gormAuthorDAO{db: db}
}
