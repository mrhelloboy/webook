package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/internal/web"
	"github.com/mrhelloboy/wehook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()

	user := initUser(db)
	user.RegisterRouters(server)

	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initWebServer() *gin.Engine {
	server := gin.Default()

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
		Build())

	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	srv := service.NewUserService(repo)
	return web.NewUserHandler(srv)
}

func initDB() *gorm.DB {
	// 数据库连接
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
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
