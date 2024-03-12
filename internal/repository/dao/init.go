package dao

import "gorm.io/gorm"

// InitTables 这种建表方式很垃圾，很不推荐。但没什么方法
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Article{})
}
