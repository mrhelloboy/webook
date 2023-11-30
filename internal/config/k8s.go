//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:123456@tcp(webook-mysql:13301)/webook?charset=utf8mb4&parseTime=True&loc=Local",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:16379",
	},
}
