package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("邮箱或者手机号码冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	Insert(ctx context.Context, u User) error
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{db: db}
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("`id` =?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error

	// 下面代码存在强耦合问题，表明是与Mysql数据库相关的
	// 如果切换成其他数据库，需要修改
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) {
		// 数据库中1062错误码通常表示“唯一性约束冲突”
		const uniqueConflictErrNo uint16 = 1062
		if mysqlError.Number == uniqueConflictErrNo {
			// 邮箱或者手机号码冲突
			return ErrUserDuplicate
		}
	}

	return err
}

// User 用户表 -> 对应数据库表结构
type User struct {
	Id int64 `gorm:"primary_key,autoIncrement"`
	// 唯一索引允许有多个空值，但不允许有多个""
	Email    sql.NullString `gorm:"unique"`
	Password string
	Phone    sql.NullString `gorm:"unique"`
	Nickname string
	Ctime    int64 // 创建时间，毫秒数
	Utime    int64 // 更新时间，毫秒数
}
