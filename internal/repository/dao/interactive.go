package dao

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

//go:generate mockgen -source=./interactive.go -package=daomocks -destination=mocks/interactive.mock.go InteractiveDAO

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error)
	DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetCollectionInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error)
}

type gormInteractiveDAO struct {
	db *gorm.DB
}

func NewGormInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &gormInteractiveDAO{db: db}
}

func (g *gormInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := g.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ? AND status = ?", biz, bizId, uid, 1).
		First(&res).Error
	return res, err
}

func (g *gormInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := g.db.WithContext(ctx).Where("biz= ? AND biz_id = ?", biz, bizId).First(&res).Error
	return res, err
}

func (g *gormInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 插入收藏记录
		err := tx.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		// 更新收藏数
		return tx.Clauses(clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
			"collect_cnt": gorm.Expr("collect_cnt + 1"),
			"utime":       now,
		})}).Create(&Interactive{
			BizId:      cb.BizId,
			Biz:        cb.Biz,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

func (g *gormInteractiveDAO) GetCollectionInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := g.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&res).Error
	return res, err
}

func (g *gormInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先记录点赞
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"status": 1,
				"utime":  now,
			}),
		}).Create(&UserLikeBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uid,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}
		// 更新点赞数
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (g *gormInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 软删除点赞记录
		err := tx.WithContext(ctx).Model(&UserLikeBiz{}).
			Where("biz = ? and biz_id = ? and uid = ?", biz, bizId, uid).
			Updates(map[string]any{
				"status": 0,
				"utime":  now,
			}).Error
		if err != nil {
			return err
		}
		// 点赞数减一
		return tx.WithContext(ctx).Model(&Interactive{}).
			Where("biz = ? and biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"like_cnt": gorm.Expr("like_cnt - 1"),
				"utime":    now,
			}).Error
	})
}

// IncrReadCnt 增加阅读量(新增或者更新）
func (g *gormInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			// 使用 SQL 表达式更新，可以解决并发问题，保证数据一致性
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		BizId:   bizId,
		Biz:     biz,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

// Interactive 互动表
type Interactive struct {
	Id int64 `gorm:"primary_key,autoIncrement"`
	// 业务标识符
	// 同一资源，在这里应该只有一行
	// 即，bizId 和 biz 要创建联合唯一索引
	BizId int64 `gorm:"uniqueIndex:biz_id_type"`
	// biz 可以是 string，也可以 int （枚举方式，但不够清晰）0-article，1-comment，2-xxx
	// 默认是 BLOB/TEXT 类型
	Biz        string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

// UserLikeBiz 用户点赞业务表
type UserLikeBiz struct {
	Id    int64  `gorm:"primary_key,autoIncrement"`
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	// 0 - 删除, 1 - 有效
	Status uint8
	Ctime  int64
	Utime  int64
}

// Collection 收藏夹 用户可以创建多个收藏夹（类似B站）
type Collection struct {
	Id    int64  `gorm:"primary_key,autoIncrement"`
	Name  string `gorm:"type=varchar(1024)"`
	Uid   int64
	Ctime int64
	Utime int64
}

// UserCollectionBiz 收藏的内容
type UserCollectionBiz struct {
	Id    int64  `gorm:"primary_key,autoIncrement"`
	Cid   int64  `gorm:"index"` // 收藏夹ID， 作为关联关系中的外键
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"uniqueIndex:biz_type_id_uid;type:varchar(128)"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}
