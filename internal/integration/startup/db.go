package startup

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db      *gorm.DB
	mongoDB *mongo.Database
)

func InitTestMongoDB() *mongo.Database {
	if mongoDB == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		moniter := &event.CommandMonitor{
			Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(moniter)
		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			panic(err)
		}
		mongoDB = client.Database("webook")
	}
	return mongoDB
}

// InitTestDB 测试的话，不用控制并发。等遇到了并发问题再说
func InitTestDB() *gorm.DB {
	if db == nil {
		dsn := "root:root@tcp(localhost:13306)/webook"
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()
			if err == nil {
				break
			}
			log.Println("等待连接 MySQL", err)
		}
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		err = dao.InitTables(db)
		if err != nil {
			panic(err)
		}
		db = db.Debug()
	}
	return db
}
