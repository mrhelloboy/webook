package ioc

import (
	"time"

	"github.com/mrhelloboy/wehook/pkg/gormx/connpool"

	promsdk "github.com/prometheus/client_golang/prometheus"

	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type (
	SrcDB *gorm.DB
	DstDB *gorm.DB
)

func InitSRC(l logger.Logger) SrcDB {
	return InitDB(l, "src")
}

func InitDST(l logger.Logger) DstDB {
	return InitDB(l, "dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB) *connpool.DoubleWritePool {
	pattern := viper.GetString("migrator.pattern")
	return connpool.NewDoubleWritePool(src.ConnPool, dst.ConnPool, pattern)
}

func InitBizDB(pool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool}))
	if err != nil {
		panic(err)
	}
	return db
}

func InitDB(l logger.Logger, key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	// 配置默认值
	cfg := Config{
		DSN: "root:root@tcp(localhost:3306)/webook-default?charset=utf8mb4&parseTime=True&loc=Local",
	}
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}

	// 数据库连接
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	cb := newCallbacks(key)
	err = db.Use(cb)
	if err != nil {
		panic(err)
	}
	// 建表
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

// sql 语句执行时间逻辑

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func (c *Callbacks) Name() string {
	return "prometheus-query"
}

func (c *Callbacks) Initialize(db *gorm.DB) error {
	c.registerAll(db)
	return nil
}

func (c *Callbacks) registerAll(db *gorm.DB) {
	// INSERT 语句
	err := db.Callback().Create().Before("*").Register("prometheus_create_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").Register("prometheus_create_after", c.after("create"))
	if err != nil {
		panic(err)
	}

	// Update 语句
	err = db.Callback().Update().Before("*").Register("prometheus_update_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").Register("prometheus_update_after", c.after("update"))
	if err != nil {
		panic(err)
	}

	// Delete 语句
	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").Register("prometheus_delete_after", c.after("delete"))
	if err != nil {
		panic(err)
	}

	// Query 语句
	err = db.Callback().Query().Before("*").Register("prometheus_query_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().After("*").Register("prometheus_query_after", c.after("query"))
	if err != nil {
		panic(err)
	}

	// Raw 语句
	err = db.Callback().Raw().Before("*").Register("prometheus_raw_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Raw().After("*").Register("prometheus_raw_after", c.after("raw"))
	if err != nil {
		panic(err)
	}

	// Row
	err = db.Callback().Row().Before("*").Register("prometheus_row_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").Register("prometheus_row_after", c.after("row"))
	if err != nil {
		panic(err)
	}
}

func (c *Callbacks) before() func(*gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}

func newCallbacks(key string) *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook_" + key,
		Name:      "gorm_query_time",
		Help:      "统计 GORM 的执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})
	c := &Callbacks{vector: vector}
	promsdk.MustRegister(vector)
	return c
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
