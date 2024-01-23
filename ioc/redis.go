package ioc

import (
	"github.com/mrhelloboy/wehook/internal/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	// redis 连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	return redisClient
}
