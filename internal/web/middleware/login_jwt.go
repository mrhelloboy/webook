package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mrhelloboy/wehook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"
)

/**
 * @Description: 登录中间件
 * Builder 模式
 */

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

// IgnorePath 忽略路径 -> 链式调用
func (l *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 排除登录和注册接口
		for _, path := range l.paths {
			log.Printf("== path: %s", path)
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 使用 jwt 校验
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 从 tokenHeader 中获取 token
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			// 格式错误 -> 可能有人在搞事
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if segs[0] != "Bearer" || segs[1] == "" {
			// 格式错误 -> 可能有人在搞事
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(segs[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"), nil
		})

		if err != nil {
			// 解析失败
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid || claims.Uid == 0 {
			// 解析失败
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 加强校验
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 有安全问题
			// todo: 写入监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// jwt token 续约 -> 比较恶心
		// 每 10 s 续约一次
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Minute))
			tokenStr, err := token.SignedString([]byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"))
			if err != nil {
				// log
				log.Println("jwt 续约失败：", err)
			}

			ctx.Header("x-jwt-token", tokenStr)
		}

		ctx.Set("claims", claims)
	}
}
