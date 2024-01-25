package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	cachemocks "github.com/mrhelloboy/wehook/internal/repository/cache/mocks"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	daomocks "github.com/mrhelloboy/wehook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	// 需要去掉毫秒之外的时间，否则在设置缓存 Set 的时候，会出错
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		// 输入
		ctx context.Context
		id  int64
		// 输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				muc := cachemocks.NewMockUserCache(ctrl)
				// 缓存未命中，返回数据不存在错误
				muc.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)

				mud := daomocks.NewMockUserDAO(ctrl)
				mud.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@gmail.com",
						Valid:  true,
					},
					Password: "hello#world123",
					Phone: sql.NullString{
						String: "18612345678",
						Valid:  true,
					},
					Nickname: "test",
					Ctime:    now.UnixMilli(),
					Utime:    now.UnixMilli(),
				}, nil)

				// 设置缓存
				muc.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@gmail.com",
					Password: "hello#world123",
					Phone:    "18612345678",
					Nickname: "test",
					Ctime:    now, //数据库存储的是毫秒数，纳秒部分被丢弃，所以这里也需要没有纳秒部分
				}).Return(nil)

				return mud, muc
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@gmail.com",
				Password: "hello#world123",
				Phone:    "18612345678",
				Nickname: "test",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "命中缓存，查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				muc := cachemocks.NewMockUserCache(ctrl)
				// 缓存命中，返回数据
				muc.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{
					Id:       123,
					Email:    "123@gmail.com",
					Password: "hello#world123",
					Phone:    "18612345678",
					Nickname: "test",
					Ctime:    now,
				}, nil)

				mud := daomocks.NewMockUserDAO(ctrl)
				return mud, muc
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@gmail.com",
				Password: "hello#world123",
				Phone:    "18612345678",
				Nickname: "test",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "缓存为命中，查询数据库出错",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				muc := cachemocks.NewMockUserCache(ctrl)
				// 缓存未命中，返回数据不存在错误
				muc.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)

				mud := daomocks.NewMockUserDAO(ctrl)
				mud.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{}, errors.New("数据库返回错误"))

				return mud, muc
			},
			ctx:      context.Background(),
			id:       123,
			wantUser: domain.User{},
			wantErr:  errors.New("数据库返回错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			cuRepo := NewUserRepository(ud, uc)
			user, err := cuRepo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
			// 为了测试协程
			// 并发测试比较困难，建议直接 review
			time.Sleep(time.Second)
		})
	}
}
