package repository

import (
	"context"

	"github.com/mrhelloboy/wehook/pkg/logger"

	"github.com/mrhelloboy/wehook/internal/domain"

	"github.com/mrhelloboy/wehook/internal/repository/cache"

	"github.com/mrhelloboy/wehook/internal/repository/dao"
)

//go:generate mockgen -source=./interactive.go -package=repomocks -destination=mocks/interactive.mock.go InteractiveRepository

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
}

type cachedInteractiveRepo struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.Logger
}

func NewCachedInteractiveRepo(dao dao.InteractiveDAO, cache cache.InteractiveCache, l logger.Logger) InteractiveRepository {
	return &cachedInteractiveRepo{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (c *cachedInteractiveRepo) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Cid:   cid,
		BizId: bizId,
		Biz:   biz,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	// 缓存收藏数
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *cachedInteractiveRepo) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 从缓存中获取阅读数、点赞数和收藏数
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	daoIntr, err := c.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	intr = c.toDomain(daoIntr)
	go func() {
		er := c.cache.Set(ctx, biz, bizId, intr)
		if er != nil {
			c.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Error(er))
		}
	}()
	return intr, nil
}

// Liked 用户是否点赞过
func (c *cachedInteractiveRepo) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, err
	default:
		return false, err
	}
}

// Collected 用户是否收藏过
func (c *cachedInteractiveRepo) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectionInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, err
	default:
		return false, err
	}
}

func (c *cachedInteractiveRepo) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 1. 插入点赞；2. 更新点赞数；3. 更新缓存
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *cachedInteractiveRepo) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *cachedInteractiveRepo) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 使用缓存记录阅读数
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 加入缓存
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *cachedInteractiveRepo) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}
