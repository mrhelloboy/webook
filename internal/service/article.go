package service

import (
	"context"

	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository/article"
	"github.com/mrhelloboy/wehook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
}

type articleSvc struct {
	authorRepo article.AuthorRepository
	readerRepo article.ReaderRepository
	l          logger.Logger
}

func NewArticleSvc(authorRepo article.AuthorRepository, readerRepo article.ReaderRepository, l logger.Logger) ArticleService {
	return &articleSvc{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
		l:          l,
	}
}

// Withdraw 撤回了帖子公开可见状态，改为私有（仅自己可见）
func (a *articleSvc) Withdraw(ctx context.Context, art domain.Article) error {
	return a.authorRepo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

// Publish 发布到线上库
// 1. 用户之前没发表过帖子，在制作库上没有记录，写完帖子直接发布
// 2. 用户之前发表过帖子，在制作库上有记录，编辑帖子再发布（更新帖子，再发布）
func (a *articleSvc) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished // 状态改为公开
	//id := art.Id
	//var err error
	//
	//// 更新或者创建制作库
	//if art.Id > 0 {
	//	err = a.authorRepo.Update(ctx, art)
	//} else {
	//	id, err = a.authorRepo.Create(ctx, art)
	//}
	//if err != nil {
	//	return 0, err
	//}
	//
	//art.Id = id
	//// 同步到线上库 - 重试
	//for i := 0; i < 3; i++ {
	//	time.Sleep(time.Second * time.Duration(i))
	//	err = a.readerRepo.Save(ctx, art)
	//	if err == nil {
	//		break
	//	}
	//	// 同步线上库失败
	//	a.l.Error("部分失败，保存到线上库失败", logger.Int64("art_id", art.Id), logger.Error(err))
	//}
	//if err != nil {
	//	a.l.Error("全部失败，重试彻底失败", logger.Int64("art_id", art.Id), logger.Error(err))
	//	// todo: 添加告警系统
	//	// todo: 或者走异步，走 Canal， MQ
	//}
	//return id, err

	// 方式二
	return a.authorRepo.Sync(ctx, art)
}

// Save 保存到制作库
func (a *articleSvc) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished // 状态改为未发布（草稿）
	if art.Id > 0 {
		err := a.authorRepo.Update(ctx, art)
		return art.Id, err
	}
	return a.authorRepo.Create(ctx, art)
}
