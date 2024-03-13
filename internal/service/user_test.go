package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository"
	repomocks "github.com/mrhelloboy/wehook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserSvc_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctl *gomock.Controller) repository.UserRepository
		// 输入
		ctx      context.Context
		email    string
		password string
		// 预期
		wantUser domain.User
		wantErr  error
	}{
		{
			name:     "登录成功",
			ctx:      context.Background(),
			email:    "123@gmail.com",
			password: "hello#world123",
			mock: func(ctl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@gmail.com").Return(domain.User{
					Id:       1,
					Email:    "123@gmail.com",
					Password: "$2a$10$EPreVaOlS89WENhOAUjOXuSFPwunL22fCJjeVQDPwfCjNwblyAcTm",
					Phone:    "18712345678",
					Nickname: "test-nick-name",
					Ctime:    now,
				}, nil)
				return userRepo
			},
			wantUser: domain.User{
				Id:       1,
				Email:    "123@gmail.com",
				Password: "$2a$10$EPreVaOlS89WENhOAUjOXuSFPwunL22fCJjeVQDPwfCjNwblyAcTm",
				Phone:    "18712345678",
				Nickname: "test-nick-name",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name:     "用户不存在",
			ctx:      context.Background(),
			email:    "123@gmail.com",
			password: "hello#world123",
			mock: func(ctl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@gmail.com").Return(domain.User{}, repository.ErrUserNotFound)
				return userRepo
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name:     "数据库查询出错",
			ctx:      context.Background(),
			email:    "123@gmail.com",
			password: "hello#world123",
			mock: func(ctl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@gmail.com").Return(domain.User{}, errors.New("error"))
				return userRepo
			},
			wantUser: domain.User{},
			wantErr:  errors.New("error"),
		},
		{
			name:     "密码错误",
			ctx:      context.Background(),
			email:    "123@gmail.com",
			password: "hell0#world123",
			mock: func(ctl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@gmail.com").Return(domain.User{
					Id:       1,
					Email:    "123@gmail.com",
					Password: "$2a$10$EPreVaOlS89WENhOAUjOXuSFPwunL22fCJjeVQDPwfCjNwblyAcTm",
					Phone:    "18712345678",
					Nickname: "test-nick-name",
					Ctime:    now,
				}, nil)
				return userRepo
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			userSvc := NewUserSvc(tc.mock(ctl), nil)
			user, err := userSvc.Login(tc.ctx, tc.email, tc.password)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestEncryptedPassword(t *testing.T) {
	password := "hello#world123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	t.Log(string(hash))
}
