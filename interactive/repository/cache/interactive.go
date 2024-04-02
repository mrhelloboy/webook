package cache

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/mrhelloboy/wehook/interactive/domain"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/interactive_incr_cnt.lua
var luaIncrCnt string

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

//go:generate mockgen -source=./interactive.go -package=cachemocks -destination=mocks/interactive.mock.go InteractiveCache

type InteractiveCache interface {
	// IncrReadCntIfPresent 如果在缓存中有对应的数据，就 +1
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	// Get 查询缓存中的数据, liked 和 collected shi不需要缓存的
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, inter domain.Interactive) error
}

type redisInteractiveCache struct {
	client redis.Cmdable
}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCache {
	return &redisInteractiveCache{client: client}
}

func (r *redisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 直接使用 HMGet，即使缓存中没有对应的 key，也不会返回 error
	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(data) == 0 {
		// 缓存不存在
		return domain.Interactive{}, ErrKeyNotExist
	}
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)

	return domain.Interactive{
		BizId:      bizId,
		ReadCnt:    readCnt,
		LikeCnt:    likeCnt,
		CollectCnt: collectCnt,
	}, nil
}

func (r *redisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, inter domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.client.HSet(ctx, key,
		fieldReadCnt, inter.ReadCnt,
		fieldLikeCnt, inter.LikeCnt,
		fieldCollectCnt, inter.CollectCnt,
	).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()
}

func (r *redisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldLikeCnt, 1).Err()
}

func (r *redisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldLikeCnt, -1).Err()
}

func (r *redisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldReadCnt, 1).Err()
}

func (r *redisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, fieldCollectCnt, 1).Err()
}

func (r *redisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
