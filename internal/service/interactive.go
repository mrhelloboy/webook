package service

import (
	"context"

	"github.com/mrhelloboy/wehook/pkg/logger"

	"github.com/mrhelloboy/wehook/internal/domain"

	"github.com/mrhelloboy/wehook/internal/repository"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService

// InteractiveService 交互服务(点赞、收藏、阅读记录等）
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
}

type interactiveSrv struct {
	interRepo repository.InteractiveRepository
	l         logger.Logger
}

func NewInteractiveService(interRepo repository.InteractiveRepository, l logger.Logger) InteractiveService {
	return &interactiveSrv{
		interRepo: interRepo,
		l:         l,
	}
}

func (i *interactiveSrv) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	// TODO implement me
	panic("implement me")
}

func (i *interactiveSrv) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	// TODO implement me
	panic("implement me")
}

// Like 点赞
func (i *interactiveSrv) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return i.interRepo.IncrLike(ctx, biz, id, uid)
}

// CancelLike 取消点赞
func (i *interactiveSrv) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.interRepo.DecrLike(ctx, biz, id, uid)
}

// IncrReadCnt 增加阅读量
func (i *interactiveSrv) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.interRepo.IncrReadCnt(ctx, biz, bizId)
}
