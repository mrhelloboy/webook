package ioc

import (
	"time"

	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	// 配置默认值
	cfg := Config{
		DSN: "root:root@tcp(localhost:3306)/webook-default?charset=utf8mb4&parseTime=True&loc=Local",
	}
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}

	// 数据库连接
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		// 添加日志
		// gormLoggerFunc 构建一个 Writer 接口
		// 日志记录器行为
		// - 慢查询阈值设置为 10 微秒（即如果查询执行时间超过10微秒，则会被记录为慢查询）
		// 一般 10ms，100ms 都算是慢查询了
		// sql 查询必须要求命中索引，最好就是走一次磁盘 IO
		// 一次磁盘 IO 是不到 10ms 的
		// - 忽略记录未找到的错误
		// - 启用参数化查询
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  glogger.Info,
		}),
	})
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

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}

// DoSomething gormLoggerFunc 涉及的技巧
type DoSomething interface {
	DoABC() string
}

type DoSomethingFunc func() string

func (d DoSomethingFunc) DoABC() string {
	return d()
}
