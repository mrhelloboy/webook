package gormx

import (
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

type DoubleWriteCallback struct {
	src     *gorm.DB
	dst     *gorm.DB
	pattern *atomicx.Value[string]
}

func (d *DoubleWriteCallback) create() func(db *gorm.DB) {
	return func(db *gorm.DB) {
	}
}
