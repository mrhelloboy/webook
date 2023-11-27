package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
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
	// 确保在序列化和反序列化的时间值时，能够正确地将其转换为字节流。
	// 这样可以保证在传输或存储时间数据时不会丢失任何信息，并且能够正确地将其还原为时间值。
	// 否则下面 sessions 获取及保存的 update_time 会出问题
	gob.Register(time.Now())
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

		// 刷新 sessions 过期时间，每 1 分钟刷新一次
		updateTime := sess.Get("update_time")
		// updateTime 为 nil，表明是登录后第一次访问
		updateTimeVal, ok := updateTime.(time.Time)
		if updateTime == nil || (ok && time.Now().Sub(updateTimeVal) > 10*time.Second) {
			sess.Set("update_time", time.Now())
			sess.Options(sessions.Options{
				MaxAge: 60 * 60 * 24, // 24 小时
			})
			if err := sess.Save(); err != nil {
				panic(err)
			}
		}
	}
}
