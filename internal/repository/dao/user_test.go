package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string
		mock func(t *testing.T) *sql.DB
		// input
		ctx  context.Context
		user User
		// output
		wantErr error
	}{
		{
			name: "插入数据成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 这边预期的是正则表达式
				// 这个写法的意思是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			ctx: context.Background(),
			user: User{
				Password: "123456",
				Phone: sql.NullString{
					String: "18612345678",
					Valid:  true,
				},
				Nickname: "test",
			},
			wantErr: nil,
		},
		{
			name: "邮箱或者手机号码冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这边预期的是正则表达式
				// 这个写法的意思是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(&mysql.MySQLError{
					Number: 1062,
				})
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: ErrUserDuplicate,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				// 这边预期的是正则表达式
				// 这个写法的意思是，只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnError(errors.New("数据库错误"))
				require.NoError(t, err)
				return mockDB
			},
			ctx:     context.Background(),
			user:    User{},
			wantErr: errors.New("数据库错误"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prepare
			mockDB := tc.mock(t)
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: mockDB,
				// 跳过版本检查 即 SELECT VERSION
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 关闭自动ping数据库以检测连接是否可用的功能
				DisableAutomaticPing: true,
				// 执行数据库操作时不开启默认事务
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			d := NewUserDAO(db)
			// execute
			err = d.Insert(tc.ctx, tc.user)
			// assert
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
