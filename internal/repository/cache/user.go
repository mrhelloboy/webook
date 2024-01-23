package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Set(ctx context.Context, user domain.User) error
	Get(ctx context.Context, id int64) (domain.User, error)
}

// RedisUserCache 用户缓存
// A 用到了 B，B 一定是接口
// A 用到了 B，B 一定是 A 的字段
// A 用到了 B, A 绝对不初始化 B，而是外面注入
type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// Get 获取用户信息
// 只要 error 为 nil， 就认为缓存里有数据
// 如果没有数据，返回一个特定的 error
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	// 如果没有数据，Redis 返回一个特定的 Redis.Nil 错误
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.key(user.Id), val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
