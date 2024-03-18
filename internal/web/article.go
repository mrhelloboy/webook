package web

import (
	"net/http"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/service"
	ijwt "github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/mrhelloboy/wehook/pkg/logger"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
)

var _ Handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc service.InteractiveService
	l        logger.Logger
	biz      string
}

func NewArticleHandler(svc service.ArticleService, interSvc service.InteractiveService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
		l:        l,
		biz:      "article",
	}
}

func (a *ArticleHandler) RegisterRouters(server *gin.Engine) {
	g := server.Group("/article")
	g.POST("/edit", a.Edit)
	g.POST("/publish", a.Publish)
	g.POST("/withdraw", a.Withdraw)
	g.POST("/list", a.List)
	g.GET("/detail/:id", a.Detail)

	// 普通用户
	pub := g.Group("/pub")
	pub.GET("/:id", a.PubDetail)
	pub.POST("/like", a.Like)
}

// Like 点赞 or 取消点赞
func (a *ArticleHandler) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req Req
	var err error
	if err = ctx.Bind(&req); err != nil {
		return
	}

	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	uc, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	if req.Like {
		err = a.interSvc.Like(ctx, a.biz, req.Id, uc.Id)
	} else {
		err = a.interSvc.CancelLike(ctx, a.biz, req.Id, uc.Id)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("点赞/取消点赞出错", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "参数错误"})
		a.l.Error("前端输入 ID 有误", logger.Error(err))
		return
	}
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "参数错误"})
		return
	}
	uc, ok := c.(*ijwt.UserClaims)
	if !ok {
		a.l.Error("未发现用户的 session 信息")
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		art, err = a.svc.GetPublishedById(ctx, id, uc.Id)
		return err
	})
	var intr domain.Interactive
	eg.Go(func() error {
		intr, err = a.interSvc.Get(ctx, a.biz, id, uc.Id)
		// 这里可以容错
		if err != nil {
			// 记日志
		}
		return err
	})

	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	// 添加阅读计数
	// 注意：因阅读记录非常频繁，并发量一大，会开启很多的 goroutine，导致有巨大压力
	//go func() {
	//	er := a.interSvc.IncrReadCnt(ctx, a.biz, art.Id)
	//	if er != nil {
	//		a.l.Error("增加阅读计数失败", logger.Int64("aid", art.Id), logger.Error(er))
	//	}
	//}()

	ctx.JSON(http.StatusOK, Result{Data: ArticleVO{
		Id:         art.Id,
		Title:      art.Title,
		Content:    art.Content,
		Status:     art.Status.ToUint8(),
		Author:     art.Author.Name,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
		LikeCnt:    intr.LikeCnt,
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
		Ctime:      art.Ctime.Format(time.DateTime),
		Utime:      art.Utime.Format(time.DateTime),
	}})
}

// Detail 获取某一帖子详情信息
func (a *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "参数错误"})
		a.l.Error("前端输入 ID 有误", logger.Error(err))
		return
	}

	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	c := ctx.MustGet("claims")
	uc, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("未发现用户的 session 信息")
		return
	}

	if art.Author.Id != uc.Id {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "输入有误"})
		a.l.Error("非法访问帖子，创作者 ID 不匹配", logger.String("访问者ID", uc.Id))
		// Todo: 需要监控上报
		return
	}

	ctx.JSON(http.StatusOK, Result{Data: ArticleVO{
		Id:    art.Id,
		Title: art.Title,
		// Abstract: art.Abstract(),
		Content: art.Content,
		Status:  art.Status.ToUint8(),
		// Author:   art.Author.Name,
		Ctime: art.Ctime.Format(time.DateTime),
		Utime: art.Utime.Format(time.DateTime),
	}})
}

// List 获取作者帖子列表(分页，只显示摘要）
func (a *ArticleHandler) List(ctx *gin.Context) {
	type Req struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	res, err := a.svc.List(ctx, claims.Id, req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVO](res, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				// Content:  src.Content,
				Status: src.Status.ToUint8(),
				// Author:   src.Author.Name,
				Ctime: src.Ctime.Format(time.DateTime),
				Utime: src.Utime.Format(time.DateTime),
			}
		}),
	})
}

// Withdraw 撤回公开发表状态的帖子，改为不可见状态
func (a *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("未发现用户的 session 信息")
		return
	}

	err := a.svc.Withdraw(ctx, domain.Article{
		Id:     req.Id,
		Author: domain.Author{Id: claims.Id},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("帖子测回失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

// Publish 帖子发布
func (a *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "请求信息错误"})
		return
	}

	// user info
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("未发现用户的 session 信息")
		return
	}

	id, err := a.svc.Publish(ctx, req.toDomain(claims.Id))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		a.l.Error("帖子发布失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{Msg: "OK", Data: id})
}

// Edit 帖子编辑
func (a *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
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

	id, err := a.svc.Save(ctx, req.toDomain(claims.Id))
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

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (r ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      r.Id,
		Title:   r.Title,
		Content: r.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
