package repository

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/repository/cache"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepo struct {
	redis *cache.RankingRedisCache
	local *cache.RankingLocalCache
}

func (c *CachedRankingRepo) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	// 先放入本地缓存，再放入redis缓存
	_ = c.local.Set(ctx, arts)
	return c.redis.Set(ctx, arts)
}

func (c *CachedRankingRepo) GetTopN(ctx context.Context) ([]domain.Article, error) {
	data, err := c.local.Get(ctx)
	if err == nil {
		return data, nil
	}
	data, err = c.redis.Get(ctx)
	if err == nil {
		_ = c.local.Set(ctx, data)
	} else {
		// redis缓存出错，从强制从本地缓存获取（不保证数据准确）
		// 为了应对redis异常时的保护措施
		return c.local.ForceGet(ctx)
	}
	return data, err
}

func NewCachedRankingRepo(redis *cache.RankingRedisCache, local *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepo{
		redis: redis,
		local: local,
	}
}
