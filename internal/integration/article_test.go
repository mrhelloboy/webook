package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhelloboy/wehook/internal/domain"

	dao "github.com/mrhelloboy/wehook/internal/repository/dao/article"
	ijwt "github.com/mrhelloboy/wehook/internal/web/jwt"

	"gorm.io/gorm"

	"github.com/mrhelloboy/wehook/internal/integration/startup"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

// SetupSuite 测试套件初始化
func (s *ArticleTestSuite) SetupSuite() {
	s.server = gin.Default()

	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})

	s.db = startup.InitTestDB()
	arthdl := startup.InitArticleHandler()
	arthdl.RegisterRouters(s.server)
}

// TearDownTest 测试套件清理
// 每一个测试方法执行后执行
func (s *ArticleTestSuite) TearDownTest() {
	// 清空所有数据，并且自增主键恢复到 1
	s.db.Exec("TRUNCATE TABLE articles")
	s.db.Exec("TRUNCATE TABLE published_articles")
}

func (s *ArticleTestSuite) TestEdit() {
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
				// 验证数据
				var art dao.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
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
				// 准备数据
				err := s.db.Create(&dao.Article{
					Id:       2,
					Title:    "新建测试帖子",
					Content:  "新建测试帖子内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(), // 已发表的状态
					Ctime:    123,
					Utime:    234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据是否更改
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
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
				// 准备数据
				err := s.db.Create(&dao.Article{
					Id:       3,
					Title:    "新建测试帖子",
					Content:  "新建测试帖子内容",
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据是否更改
				var art dao.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
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
			assert.Equal(t, tc.wantRes, webRes)

			tc.after(t)
		})
	}
}

func (s *ArticleTestSuite) TestPublish() {
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

				// 制作库验证
				var art dao.Article
				err := s.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				// 确保生成主键
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "test title",
					Content:  "test content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
				}, art)

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.db.Where("author_id = ?", 123).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Id = 0
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{Article: dao.Article{
					Title:    "test title",
					Content:  "test content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
				}}, publishedArt)
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
				// 模拟已有存在的帖子
				err := s.db.Create(&dao.Article{
					Id:       2,
					Title:    "test title",
					Content:  "test content",
					AuthorId: 123,
					Ctime:    456,
					Utime:    234,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据

				// 制作库验证
				var art dao.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				// 更新时间发生变化
				assert.True(t, art.Utime > 234)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "new test title",
					Content:  "new test content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
					Ctime:    456,
				}, art)

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.db.Where("id = ?", 2).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{Article: dao.Article{
					Id:       2,
					Title:    "new test title",
					Content:  "new test content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
				}}, publishedArt)
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
				// 模拟已有制作库和线上库存在的帖子
				art := dao.Article{
					Id:       3,
					Title:    "test title",
					Content:  "test content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    234,
					Utime:    345,
				}
				err := s.db.Create(&art).Error
				assert.NoError(t, err)

				pubArt := dao.PublishedArticle{Article: art}
				err = s.db.Create(&pubArt).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据

				// 制作库验证
				var art dao.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				// 更新时间发生变化
				assert.True(t, art.Utime > 345)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "new test title",
					Content:  "new test content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
					Ctime:    234,
				}, art)

				// 线上库验证
				var publishedArt dao.PublishedArticle
				err = s.db.Where("id = ?", 3).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Utime > 345)
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{Article: dao.Article{
					Id:       3,
					Title:    "new test title",
					Content:  "new test content",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
				}}, publishedArt)
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
				err := s.db.Create(&art).Error
				assert.NoError(t, err)

				pubArt := dao.PublishedArticle{Article: art}
				err = s.db.Create(&pubArt).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据 （数据没有发生变更)

				// 制作库验证
				var art dao.Article
				err := s.db.Where("id = ?", 4).First(&art).Error
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
				err = s.db.Where("id = ?", 4).First(&publishedArt).Error
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
			assert.Equal(t, tc.wantRes, webRes)

			tc.after(t)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
