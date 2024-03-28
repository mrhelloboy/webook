package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	// 配置默认值
	cfg := Config{
		Addr: "localhost:6379",
	}
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}
	// redis 连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	return redisClient
}

// InitRLockClient 初始化分布式锁客户端
func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
