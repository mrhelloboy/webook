package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

/**
 * @Description: 登录中间件
 * Builder 模式
 */

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

// IgnorePath 忽略路径 -> 链式调用
func (l *LoginMiddlewareBuilder) IgnorePath(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 排除登录和注册接口
		for _, path := range l.paths {
			log.Printf("== path: %s", path)
			if ctx.Request.URL.Path == path {
				return
			}
		}

		//if ctx.Request.URL.Path == "/user/login" || ctx.Request.URL.Path == "/user/signup" {
		//	return
		//}

		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
