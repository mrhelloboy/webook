package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler() *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRouters(server *gin.Engine) {
	server.POST("/user/login", u.Login)
	server.POST("/user/signup", u.SignUp)
	server.POST("/user/edit", u.Edit)
	server.GET("/user/profile", u.Profile)
}

// Login 登录用户
func (u *UserHandler) Login(ctx *gin.Context) {

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

	ctx.String(http.StatusOK, "注册成功")
}

// Edit 修改用户信息
func (u *UserHandler) Edit(ctx *gin.Context) {

}

// Profile 用户信息
func (u *UserHandler) Profile(ctx *gin.Context) {

}