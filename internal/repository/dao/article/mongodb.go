package article

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoDBAuthorDAO struct {
	col           *mongo.Collection // 制作库
	liveCol       *mongo.Collection // 线上库
	snowflakeNode *snowflake.Node
	// idGen         IDGenerator
}

func (m *mongoDBAuthorDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mongoDBAuthorDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mongoDBAuthorDAO) GetById(ctx context.Context, id int64) (Article, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mongoDBAuthorDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mongoDBAuthorDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	id := m.snowflakeNode.Generate().Int64()
	art.Id = id
	_, err := m.col.InsertOne(ctx, art)
	return id, err
}

// UpdateById 更新制作库
func (m *mongoDBAuthorDAO) UpdateById(ctx context.Context, art Article) error {
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (m *mongoDBAuthorDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 没法引入事务处理
	id := art.Id
	var err error
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	// 操作线上库
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now
	update := bson.M{
		"$set": PublishedArticle{
			Article: art,
		},
		"$setOnInsert": bson.M{"ctime": now},
	}
	filter := bson.M{"id": id}
	_, err = m.liveCol.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return id, err
}

func (m *mongoDBAuthorDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	// TODO implement me
	panic("implement me")
}

// InitCollection 初始化集合及索引
func InitCollection(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{bson.E{Key: "author_id", Value: 1}, bson.E{Key: "ctime", Value: 1}},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().CreateMany(ctx, index)
	return err
}

func NewMongoDBAuthorDAO(db *mongo.Database, node *snowflake.Node) AuthorDAO {
	return &mongoDBAuthorDAO{
		col:           db.Collection("articles"),
		liveCol:       db.Collection("published_articles"),
		snowflakeNode: node,
	}
}

//type IdGenerator func() int64
//
//func NewMongoDBAuthorDAOV1(db *mongo.Database, idGen IdGenerator) AuthorDAO {
//	return &mongoDBAuthorDAO{
//		col:     db.Collection("articles"),
//		liveCol: db.Collection("published_articles"),
//		idGen:   idGen,
//	}
//}
