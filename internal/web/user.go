package web

import (
	"errors"
	"net/http"

	"github.com/mrhelloboy/wehook/internal/errs"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/service"
	myjwt "github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const biz = "login"

// 确保 UserHandler 实现了 Handler 接口
var _ Handler = (*UserHandler)(nil)

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	phoneExp    *regexp.Regexp
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	cmd         redis.Cmdable
	myjwt.Handler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, cmd redis.Cmdable, jwtHandler myjwt.Handler) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		phoneRegexPattern    = `^1[0-9]{10}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	phoneExp := regexp.MustCompile(phoneRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
		phoneExp:    phoneExp,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		cmd:         cmd,
		Handler:     jwtHandler,
	}
}

func (u *UserHandler) RegisterRouters(server *gin.Engine) {
	ug := server.Group("/user")
	ug.POST("/login", u.Login)
	ug.POST("/loginJWT", u.LoginJWT)
	ug.POST("/logout", u.Logout)
	ug.POST("/logoutJWT", u.LogoutJWT)
	ug.POST("/signup", u.SignUp)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
	ug.GET("/profileJWT", u.ProfileJWT)

	ug.POST("/login_sms/code/send", u.SendLoginSmsCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/refresh_token", u.RefreshToken)
}

// RefreshToken 用于刷新 JWT token
// 为了安全性，可以同时刷新长短 token，用 Redis 来记录是否有效，即 refresh_token 是一次性的。
// 及参考登录校验中，比较 User-Agent 来增强安全性。
func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	// 从这个接口获取的 token 是 refresh_token
	refreshToken := u.ExtractToken(ctx)
	var rc myjwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return myjwt.RtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// Redis 出问题或者用户已经退出登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 设置新的 JWT token
	err = u.SetJWTToken(ctx, rc.Id, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		zap.L().Error("设置 JWT token 失败", zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "ok"})
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 验证手机号是否合法
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "手机号码不合法"})
		return
	}

	ok, err = u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		zap.L().Error("校验验证码出错", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码错误"})
		return
	}

	// 查找或者创建用户
	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "通过手机号登录成功"})
}

func (u *UserHandler) SendLoginSmsCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 验证手机号是否合法
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "手机号码不合法"})
		return
	}

	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{Msg: "发送太频繁，请稍后再试"})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

// Login 登录用户
func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "用户名或者密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 设置 sessions
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	// todo: Secure 和 HttpOnly 要在生产环境开启
	//sess.Options(sessions.Options{
	//	Secure:   true,
	//	HttpOnly: true,
	//})
	_ = sess.Save()

	ctx.String(http.StatusOK, "登录成功")
	return
}

// LoginJWT 使用JWT方式登录用户
func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, Result{Code: errs.UserInvalidOrPassword, Msg: "用户名不存在或者密码错误"})
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 用 JWT 设置登录态
	if err := u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	ctx.JSON(http.StatusOK, Result{Code: 2, Msg: "使用JWT登录成功"})
	return
}

// Logout 退出登录, 清除用户登录状态所保存的相关信息
func (u *UserHandler) Logout(ctx *gin.Context) {
	// session 处理方案
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		// Secure: true,
		// HttpOnly: true,
		MaxAge: -1, // 立即过期
	})
	sess.Clear()
	_ = sess.Save()
	ctx.String(http.StatusOK, "退出成功")
	return
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "退出登录失败"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "退出成功"})
}

// SignUp 注册用户
func (u *UserHandler) SignUp(ctx *gin.Context) {
	// SignUpReq 放在方法内，是因为 SignUpReq 信息只和注册用户有关，没必要给其他方法使用
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 自动解析数据到 req 中
	// 解析出错，会返回 400 错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验，不论前端有没有校验，后端也要校验
	// 邮箱格式校验
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}

	// 密码校验
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}

	// 调用 Service 层的注册方法
	err = u.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if errors.Is(err, service.ErrUserDuplicate) {
		ctx.String(http.StatusOK, "邮箱已存在, 请换一个")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

// Edit 修改用户信息（手机、邮箱、密码的修改需要验证才能修改）
func (u *UserHandler) Edit(ctx *gin.Context) {
}

// Profile 用户信息
func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "这是你的Profile")
}

// ProfileJWT 用户信息
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		Nickname string
	}
	uc := ctx.MustGet("claims").(*myjwt.UserClaims)
	user, err := u.svc.Profile(ctx, uc.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: Profile{
			Email:    user.Email,
			Phone:    user.Phone,
			Nickname: user.Nickname,
		},
		Code: 2,
		Msg:  "ok",
	})
}
