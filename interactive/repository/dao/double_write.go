package dao

import (
	"context"
	"errors"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

const (
	patternDstOnly  = "DST_ONLY"
	patternSrcOnly  = "SRC_ONLY"
	patternDstFirst = "DST_FIRST"
	patternSrcFirst = "SRC_FIRST"
)

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string]
}

func NewDoubleWriteDAOV1(src *gorm.DB, dst *gorm.DB) *DoubleWriteDAO {
	return &DoubleWriteDAO{
		src:     NewGormInteractiveDAO(src),
		dst:     NewGormInteractiveDAO(dst),
		pattern: atomicx.NewValueOf(patternDstOnly),
	}
}

func NewDoubleWriteDAO(src InteractiveDAO, dst InteractiveDAO) *DoubleWriteDAO {
	return &DoubleWriteDAO{
		src:     src,
		dst:     dst,
		pattern: atomicx.NewValueOf(patternDstOnly),
	}
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case patternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// log
			// dst 写失败，不被认为是失败
		}
		return nil
	case patternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	case patternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// log
		}
		return nil
	default:
		return errors.New("未知的双写模式")
	}
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	// TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	// TODO implement me
	panic("implement me")
}
