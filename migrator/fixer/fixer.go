package fixer

import (
	"context"
	"errors"

	"github.com/mrhelloboy/wehook/migrator"
	"github.com/mrhelloboy/wehook/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Fixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func (f *Fixer[T]) Fix(ctx context.Context, evt events.InconsistentEvent) error {
	var t T
	err := f.base.WithContext(ctx).Where("id=?", evt.ID).First(&t).Error
	switch err {
	case nil:
		// base有数据
		// 修复数据时，可以考虑增加 WHERE base.utime >= target.utime
		// 如果 utime用不了，就看有没有version之类的，或者能够判定数据新老的
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	case gorm.ErrRecordNotFound:
		return f.target.WithContext(ctx).Where("id=?", evt.ID).Delete(&t).Error
	default:
		return err
	}
}

func (f *Fixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing, events.InconsistentEventTypeNEQ:
		var t T
		err := f.base.WithContext(ctx).Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return f.target.WithContext(ctx).Where("id=?", evt.ID).Delete(new(T)).Error
		case nil:
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.WithContext(ctx).Where("id=?", evt.ID).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}

func (f *Fixer[T]) FixV2(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.WithContext(ctx).Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return nil
		case nil:
			return f.target.Create(&t).Error
		default:
			return nil
		}
	case events.InconsistentEventTypeNEQ:
		var t T
		err := f.base.WithContext(ctx).Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return f.target.WithContext(ctx).Where("id=?", evt.ID).Delete(&t).Error
		case nil:
			return f.target.Updates(&t).Error
		default:
			return nil
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.WithContext(ctx).Where("id=?", evt.ID).Delete(new(T)).Error
	default:
		return errors.New("未知的不一致类型")
	}
}
