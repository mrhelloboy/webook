//go:build e2e

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bwmarrin/snowflake"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mrhelloboy/wehook/internal/domain"

	dao "github.com/mrhelloboy/wehook/internal/repository/dao/article"
	ijwt "github.com/mrhelloboy/wehook/internal/web/jwt"

	"github.com/mrhelloboy/wehook/internal/integration/startup"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"
)

// ArticleMongoHandlerTestSuite 测试套件
type ArticleMongoHandlerTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

// SetupSuite 测试套件初始化
func (s *ArticleMongoHandlerTestSuite) SetupSuite() {
	s.server = gin.Default()

	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Id: 123,
		})
	})

	s.mdb = startup.InitTestMongoDB()
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	err = dao.InitCollection(s.mdb)
	if err != nil {
		panic(err)
	}
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")

	arthdl := startup.InitArticleHandler(dao.NewMongoDBAuthorDAO(s.mdb, node))
	arthdl.RegisterRouters(s.server)
}

// TearDownTest 测试套件清理
// 每一个测试方法执行后执行
func (s *ArticleMongoHandlerTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.mdb.Collection("articles").DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoHandlerTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		// 预期中的输入
		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证数据
				var art dao.Article
				// err := s.db.Where("id=?", 1).First(&art).Error
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "新建测试帖子",
					Content:  "新建测试帖子内容",
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "新建测试帖子",
				Content: "新建测试帖子内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			name: "修改已有帖子，并保存成功",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				// 准备数据
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:       2,
					Title:    "新建测试帖子",
					Content:  "新建测试帖子内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(), // 已发表的状态
					Ctime:    123,
					Utime:    234,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证数据是否更改
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Utime > 234)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新建测试帖子-2",
					Content:  "新建测试帖子内容-2",
					Status:   domain.ArticleStatusUnpublished.ToUint8(), // 编辑过后变为草稿状态
					Ctime:    123,
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新建测试帖子-2",
				Content: "新建测试帖子内容-2",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name: "修改他人的帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 准备数据
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:       3,
					Title:    "新建测试帖子",
					Content:  "新建测试帖子内容",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证数据是否更改
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				// 数据没有变化
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "新建测试帖子",
					Content:  "新建测试帖子内容",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新建测试帖子-2",
				Content: "新建测试帖子内容-2",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			// 构造请求
			reqBody, err := json.Marshal(tc.art)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/article/edit", bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 执行
			s.server.ServeHTTP(resp, req)
			// 验证结果
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}

			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			// 只能判断有ID，因雪花算法无法确定具体的值
			if tc.wantRes.Data > 0 {
				assert.True(t, webRes.Data > 0)
			}

			tc.after(t)
		})
	}
}

func (s *ArticleMongoHandlerTestSuite) TestPublish() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		// 预期中的输入
		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				// 验证数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 制作库验证
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				assert.NoError(t, err)
				// 确保生成主键
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.Equal(t, "test title", art.Title)
				assert.Equal(t, "test content", art.Content)
				assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				assert.Equal(t, int64(123), art.AuthorId)

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				assert.Equal(t, "test title", publishedArt.Title)
				assert.Equal(t, "test content", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
			},
			art: Article{
				Title:   "test title",
				Content: "test content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			name: "更新未发表的帖子并发表", // 制作库上有，但线上库上没有的情况
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已有存在的帖子
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:       2,
					Title:    "test title",
					Content:  "test content",
					AuthorId: 123,
					Ctime:    456,
					Utime:    234,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 制作库验证
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				// 更新时间发生变化
				assert.True(t, art.Utime > 234)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, "new test title", art.Title)
				assert.Equal(t, "new test content", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.True(t, publishedArt.Utime > 234)
				assert.True(t, publishedArt.Ctime > 0)
				assert.Equal(t, "new test title", publishedArt.Title)
				assert.Equal(t, "new test content", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
			},
			art: Article{
				Id:      2,
				Title:   "new test title",
				Content: "new test content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name: "更新帖子，并重新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已有存在的帖子
				art := dao.Article{
					Id:       3,
					Title:    "test title",
					Content:  "test content",
					AuthorId: 123,
					Ctime:    456,
					Utime:    234,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}
				_, err := s.col.InsertOne(ctx, art)
				assert.NoError(t, err)

				pubArt := dao.PublishedArticle{Article: art}
				_, err = s.liveCol.InsertOne(ctx, pubArt)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证数据
				// 制作库验证
				var art dao.Article
				// err := s.db.Where("id = ?", 3).First(&art).Error
				err := s.col.FindOne(ctx, bson.M{"id": 3}).Decode(&art)
				assert.NoError(t, err)
				// 更新时间发生变化
				assert.True(t, art.Utime > 234)
				assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				assert.Equal(t, "new test title", art.Title)
				assert.Equal(t, "new test content", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(123), art.AuthorId)

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"id": 3}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.True(t, publishedArt.Utime > 345)
				assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), publishedArt.Status)
				assert.Equal(t, "new test title", publishedArt.Title)
				assert.Equal(t, "new test content", publishedArt.Content)
				assert.Equal(t, int64(456), publishedArt.Ctime)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
			},
			art: Article{
				Id:      3,
				Title:   "new test title",
				Content: "new test content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 3,
			},
		},
		{
			name: "更新他人帖子，及发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 模拟已有制作库和线上库存在的帖子
				art := dao.Article{
					Id:       4,
					Title:    "test title",
					Content:  "test content",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    234,
					Utime:    345,
				}
				_, err := s.col.InsertOne(ctx, art)
				assert.NoError(t, err)

				pubArt := dao.PublishedArticle{Article: art}
				_, err = s.liveCol.InsertOne(ctx, pubArt)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据 （数据没有发生变更)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 制作库验证
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&art)
				assert.NoError(t, err)
				// 数据没有发生变化
				assert.Equal(t, art.Title, "test title")
				assert.Equal(t, art.Content, "test content")
				assert.Equal(t, art.AuthorId, int64(789))
				assert.Equal(t, art.Ctime, int64(234))
				assert.Equal(t, art.Utime, int64(345))
				assert.Equal(t, art.Status, domain.ArticleStatusPublished.ToUint8())

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, publishedArt.Title, "test title")
				assert.Equal(t, publishedArt.Content, "test content")
				assert.Equal(t, publishedArt.AuthorId, int64(789))
				assert.Equal(t, publishedArt.Ctime, int64(234))
				assert.Equal(t, publishedArt.Utime, int64(345))
				assert.Equal(t, publishedArt.Status, domain.ArticleStatusPublished.ToUint8())
			},
			art: Article{
				Id:      4,
				Title:   "new test title",
				Content: "new test content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "系统错误",
				Code: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			// 构造请求
			reqBody, err := json.Marshal(tc.art)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/article/publish", bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			// 执行
			s.server.ServeHTTP(resp, req)
			// 验证结果
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)

			if tc.wantRes.Data > 0 {
				assert.True(t, webRes.Data > 0)
			}

			tc.after(t)
		})
	}
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, &ArticleMongoHandlerTestSuite{})
}
