package fixer

import (
	"context"

	"github.com/mrhelloboy/wehook/pkg/migrator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	var t T
	rows, err := base.Model(&t).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

func (o *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var src T
	err := o.base.WithContext(ctx).Where("id=?", id).First(&src).Error
	switch err {
	case nil:
		return o.target.Clauses(&clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(o.columns),
		}).Create(&src).Error
	case gorm.ErrRecordNotFound:
		return o.target.Delete("id=?", id).Error
	default:
		return err
	}
}
