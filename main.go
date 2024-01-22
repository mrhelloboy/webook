package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/config"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/internal/service/sms/memory"
	"github.com/mrhelloboy/wehook/internal/web"
	"github.com/mrhelloboy/wehook/internal/web/middleware"
	"github.com/mrhelloboy/wehook/pkg/ginx/middlewares/ratelimit"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	redisClient := initRedis()
	server := initWebServer(redisClient)

	user := initUser(db, redisClient)
	user.RegisterRouters(server)

	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.String(200, "hello world")
	//})
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initWebServer(redisClient redis.Cmdable) *gin.Engine {
	server := gin.Default()

	// redis 客户端
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr:     config.Config.Redis.Addr,
	//	Password: "", // no password set
	//	DB:       1,  // use default DB
	//})

	// 限流
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	// 跨域中间件
	server.Use(cors.New(cors.Config{
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许跨域使用的 header，否则前端无法读取 x-jwt-token
		// 前端读取 x-jwt-token 的值来配置 Authorization 头
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// 开发环境允许跨域
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your-company.com")
		},
		MaxAge: 1 * time.Hour,
	}))

	// sessions 中间件

	// 不建议使用 Cookie 方式存放 session
	//store := cookie.NewStore([]byte("secret"))

	// 单体应用可以使用内存方式存放 session。参数说明：
	// 第一个为：authentication key:
	// 第二个为：encryption key:
	// store := memstore.NewStore([]byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"), []byte("gJbN8K6q2sUc2PVbDM3DfJwYNrrHqmXg"))

	// 分布式应用建议用 Redis 方式存放 session。参数说明：
	// size: 最大空闲连接数.
	// network: 一般都是 tcp
	// address: host:port
	// password: redis-password
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"),
	//	[]byte("gJbN8K6q2sUc2PVbDM3DfJwYNrrHqmXg"))
	//if err != nil {
	//	panic(err)
	//}
	//server.Use(sessions.Sessions("ssid", store))
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgnorePath("/user/signup").
	//	IgnorePath("/user/login").
	//	Build())

	// 使用 jwt
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePath("/user/signup").
		IgnorePath("/user/loginJWT").
		IgnorePath("/user/login_sms").
		IgnorePath("/user/login_sms/code/send").
		Build())

	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, uc)
	userSvc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)

	return web.NewUserHandler(userSvc, codeSvc)
}

func initDB() *gorm.DB {
	// 数据库连接
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
	if err != nil {
		// 只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	// 建表
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initRedis() *redis.Client {
	// redis 连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	return redisClient
}
