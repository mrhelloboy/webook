package cache

import (
	"context"
	"errors"
	"github.com/mrhelloboy/wehook/internal/repository/cache/redismocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable
		// 输入
		ctx   context.Context
		biz   string
		phone string
		code  string
		// 预期
		wantErr error
	}{
		{
			name: "验证码设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				mc := redismocks.NewMockCmdable(ctrl)

				// Redis.Cmd 对象
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(0))

				mc.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:18612345678"}, "123456").Return(cmd)
				return mc
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "18612345678",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "Redis 执行 Lua 脚本获取结果失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				mc := redismocks.NewMockCmdable(ctrl)

				// Redis.Cmd 对象
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis error"))

				mc.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:18612345678"}, "123456").Return(cmd)
				return mc
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "18612345678",
			code:    "123456",
			wantErr: errors.New("redis error"),
		},
		{
			name: "验证码发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				mc := redismocks.NewMockCmdable(ctrl)

				// Redis.Cmd 对象
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-1))

				mc.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:18612345678"}, "123456").Return(cmd)
				return mc
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "18612345678",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				mc := redismocks.NewMockCmdable(ctrl)

				// Redis.Cmd 对象
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-2))

				mc.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:18612345678"}, "123456").Return(cmd)
				return mc
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "18612345678",
			code:    "123456",
			wantErr: errors.New("系统错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cc := NewCodeCache(tc.mock(ctrl))
			err := cc.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
