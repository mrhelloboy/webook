package service

import (
	"context"

	events "github.com/mrhelloboy/wehook/internal/events/article"

	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository/article"
	"github.com/mrhelloboy/wehook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleSvc struct {
	authorRepo article.AuthorRepository
	l          logger.Logger
	producer   events.Producer
}

func NewArticleSvc(authorRepo article.AuthorRepository, l logger.Logger, producer events.Producer) ArticleService {
	return &articleSvc{
		authorRepo: authorRepo,
		l:          l,
		producer:   producer,
	}
}

func (a *articleSvc) GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	art, err := a.authorRepo.GetPublishedById(ctx, id)

	if err == nil {
		go func() {
			// 使用消息队列，发送阅读事件，增加阅读数计数
			er := a.producer.ProduceReadEvent(ctx, events.ReadEvent{
				// 即便消费者要用 art 里面的数据，
				// 应该让它去查询，不要在 event 里面带
				Uid: uid,
				Aid: id,
			})
			if er != nil {
				a.l.Error("发送读者阅读事件失败")
			}
		}()
	}
	return art, err
}

func (a *articleSvc) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.authorRepo.GetById(ctx, id)
}

func (a *articleSvc) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.authorRepo.List(ctx, uid, offset, limit)
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
