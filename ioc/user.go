package ioc

import (
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/pkg/redisx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// InitUserCache 配合 PrometheusHook 使用
func InitUserCache(client *redis.ClusterClient) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "user_cache",
		Help:      "统计 User 的 缓存命中情况",
		ConstLabels: map[string]string{
			"biz": "user",
		},
	}))
	panic("别调用")
}
