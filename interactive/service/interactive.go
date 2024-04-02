package service

import (
	"context"

	"github.com/mrhelloboy/wehook/interactive/domain"
	"github.com/mrhelloboy/wehook/interactive/repository"
	"github.com/mrhelloboy/wehook/pkg/logger"

	"golang.org/x/sync/errgroup"
)

// InteractiveService 交互服务 ( 点赞、收藏、阅读记录等）
//
//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
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

func (i *interactiveSrv) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := i.interRepo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (i *interactiveSrv) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	return i.interRepo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveSrv) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	var eg errgroup.Group
	var intr domain.Interactive
	var liked bool
	var collected bool

	eg.Go(func() error {
		var err error
		intr, err = i.interRepo.Get(ctx, biz, bizId)
		return err
	})
	// 是否点赞过
	eg.Go(func() error {
		var err error
		liked, err = i.interRepo.Liked(ctx, biz, bizId, uid)
		return err
	})
	// 是否收藏过
	eg.Go(func() error {
		var err error
		collected, err = i.interRepo.Collected(ctx, biz, bizId, uid)
		return err
	})
	err := eg.Wait()
	if err != nil {
		return domain.Interactive{}, err
	}
	intr.Liked = liked
	intr.Collected = collected
	return intr, nil
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
