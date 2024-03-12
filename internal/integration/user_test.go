package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mrhelloboy/wehook/internal/integration/startup"
	"github.com/mrhelloboy/wehook/internal/web"
	"github.com/mrhelloboy/wehook/ioc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_e2e_SendLoginSmsCode(t *testing.T) {
	server := startup.InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string
		// 准备数据
		before func(t *testing.T)
		// 验证数据
		after        func(t *testing.T)
		reqBody      string
		wantCode     int
		wantRespBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				v, err := rdb.GetDel(ctx, "phone_code:login:18612345678").Result()
				cancel()
				assert.NoError(t, err)
				// 验证验证码是否为 6 位
				assert.True(t, len(v) == 6)
			},
			reqBody:      `{"phone":"18612345678"}`,
			wantCode:     200,
			wantRespBody: web.Result{Msg: "发送成功"},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				_, err := rdb.Set(ctx, "phone_code:login:18612345678", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				v, err := rdb.GetDel(ctx, "phone_code:login:18612345678").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", v)
			},
			reqBody:      `{"phone":"18612345678"}`,
			wantCode:     200,
			wantRespBody: web.Result{Msg: "发送太频繁，请稍后再试"},
		},
		{
			name: "系统错误，没有设置过期时间",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				_, err := rdb.Set(ctx, "phone_code:login:18612345678", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				v, err := rdb.GetDel(ctx, "phone_code:login:18612345678").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", v)
			},
			reqBody:      `{"phone":"18612345678"}`,
			wantCode:     200,
			wantRespBody: web.Result{Code: 5, Msg: "系统错误"},
		},
		{
			name: "手机号码为空",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody:      `{"phone":""}`,
			wantCode:     200,
			wantRespBody: web.Result{Code: 4, Msg: "手机号码不合法"},
		},
		{
			name:     "数据格式错误",
			before:   func(t *testing.T) {},
			after:    func(t *testing.T) {},
			reqBody:  `{"phone":}`,
			wantCode: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			// request
			req, err := http.NewRequest(http.MethodPost, "/user/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			// response
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}

			var res web.Result
			// err = json.Unmarshal(resp.Body.Bytes(), &res)
			err = json.NewDecoder(resp.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRespBody, res)

			tc.after(t)
		})
	}
}
