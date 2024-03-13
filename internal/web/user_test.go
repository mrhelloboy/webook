package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/service"
	svcmocks "github.com/mrhelloboy/wehook/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func(ctrl *gomock.Controller) service.UserService
		reqBody      string
		wantCode     int
		wantRespBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello@world123",
				}).Return(nil)
				return usersvc
			},
			reqBody: `
			{
				"email":"123@qq.com",
				"password":"hello@world123",
				"confirmPassword":"hello@world123"
			}`,
			wantCode:     http.StatusOK,
			wantRespBody: "注册成功",
		},
		{
			name: "参数错误，bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
			{
				"email": "123@qq.com",
				"password": "hello@world123"
			`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
			{
				"email":"123@qq",
				"password":"hello@world123",
				"confirmPassword":"hello@world123"
			}`,
			wantCode:     http.StatusOK,
			wantRespBody: "邮箱格式错误",
		},
		{
			name: "两次密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
			{
				"email":"123@qq.com",
				"password":"hello@world123",
				"confirmPassword":"Hello@world123"
			}`,
			wantCode:     http.StatusOK,
			wantRespBody: "两次密码不一致",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
			{
				"email":"123@qq.com",
				"password":"helloworld123",
				"confirmPassword":"helloworld123"
			}`,
			wantCode:     http.StatusOK,
			wantRespBody: "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
		},
		{
			name: "注册邮箱已存在",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello@world123",
				}).Return(service.ErrUserDuplicate)
				return usersvc
			},
			reqBody: `
			{
				"email":"123@qq.com",
				"password":"hello@world123",
				"confirmPassword":"hello@world123"
			}`,
			wantCode:     http.StatusOK,
			wantRespBody: "邮箱已存在, 请换一个",
		},
		{
			name: "调用Signup失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello@world123",
				}).Return(errors.New("未知错误"))
				return usersvc
			},
			reqBody: `
			{
				"email":"123@qq.com",
				"password":"hello@world123",
				"confirmPassword":"hello@world123"
			}`,
			wantCode:     http.StatusOK,
			wantRespBody: "系统异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil, nil, nil)
			h.RegisterRouters(server)

			// 构建请求
			req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			// 构建响应
			resp := httptest.NewRecorder()

			// HTTP 请求进去 GIN 框架的入口
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantRespBody, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBody      string
		wantCode     int
		wantRespBody func() string
	}{
		{
			name: "通过手机号码登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(), "login", "18612345678", "123456").Return(true, nil)
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().FindOrCreate(gomock.Any(), "18612345678").Return(domain.User{Id: 1}, nil)
				return usersvc, codesvc
			},
			reqBody:  `{"phone":"18612345678", "code":"123456"}`,
			wantCode: http.StatusOK,
			wantRespBody: func() string {
				d, _ := json.Marshal(Result{Code: 4, Msg: "通过手机号登录成功"})
				return string(d)
			},
		},
		{
			name: "请求参数异常，Bind失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc, codesvc
			},
			reqBody:  `{"phone":"18612345678", "code":"123456"`,
			wantCode: http.StatusBadRequest,
			wantRespBody: func() string {
				return ""
			},
		},
		{
			name: "手机号码不合法",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc, codesvc
			},
			reqBody:  `{"phone":"1861234567", "code":"123456"}`,
			wantCode: http.StatusOK,
			wantRespBody: func() string {
				d, _ := json.Marshal(Result{Code: 4, Msg: "手机号码不合法"})
				return string(d)
			},
		},
		{
			name: "短信验证码验证异常",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(), "login", "18612345678", "123456").Return(true, errors.New("系统错误"))
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc, codesvc
			},
			reqBody:  `{"phone":"18612345678", "code":"123456"}`,
			wantCode: http.StatusOK,
			wantRespBody: func() string {
				d, _ := json.Marshal(Result{Code: 5, Msg: "系统错误"})
				return string(d)
			},
		},
		{
			name: "短信验证码错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(), "login", "18612345678", "123456").Return(false, nil)
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc, codesvc
			},
			reqBody:  `{"phone":"18612345678", "code":"123456"}`,
			wantCode: http.StatusOK,
			wantRespBody: func() string {
				d, _ := json.Marshal(Result{Code: 4, Msg: "验证码错误"})
				return string(d)
			},
		},
		{
			name: "查找或创建用户失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(), "login", "18612345678", "123456").Return(true, nil)
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().FindOrCreate(gomock.Any(), "18612345678").Return(domain.User{Id: 1}, errors.New("error"))
				return usersvc, codesvc
			},
			reqBody:  `{"phone":"18612345678", "code":"123456"}`,
			wantCode: http.StatusOK,
			wantRespBody: func() string {
				d, _ := json.Marshal(Result{Code: 5, Msg: "系统错误"})
				return string(d)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// gin server and user handler
			server := gin.Default()
			usersvc, codesvc := tc.mock(ctrl)
			h := NewUserHandler(usersvc, codesvc, nil, nil)
			h.RegisterRouters(server)
			// request
			req, err := http.NewRequest(http.MethodPost, "/user/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			// response
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantRespBody(), resp.Body.String())
		})
	}
}
