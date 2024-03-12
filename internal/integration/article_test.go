package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhelloboy/wehook/internal/repository/dao"
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

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
