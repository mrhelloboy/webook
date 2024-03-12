package web

import (
	"net/http"

	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/service"
	ijwt "github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/mrhelloboy/wehook/pkg/logger"

	"github.com/gin-gonic/gin"
)

var _ Handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.Logger
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (a *ArticleHandler) RegisterRouters(server *gin.Engine) {
	g := server.Group("/article")
	g.POST("/edit", a.Edit)
}

// Edit 帖子编辑
func (a *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "请求信息错误"})
		return
	}

	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}

	id, err := a.svc.Save(ctx, domain.Article{
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("保存帖子失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{Msg: "OK", Data: id})
}
