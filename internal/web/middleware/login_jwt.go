package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mrhelloboy/wehook/internal/web"
	"log"
	"net/http"
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

		tokenStr := web.ExtractToken(ctx)
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"), nil
		})

		if err != nil {
			// 解析失败
			log.Println("jwt 解析失败：", err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid || claims.Uid == 0 {
			// 解析失败
			log.Println("token valid fail or claims.uid == 0")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 加强校验
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 有安全问题
			// todo: 写入监控
			log.Printf("claims.UserAgent (%s) != ctx.Request.UserAgent() %s\n", claims.UserAgent, ctx.Request.UserAgent())
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// jwt token 续约 -> 比较恶心
		// 每 1 分钟续约一次
		//now := time.Now()
		//if claims.ExpiresAt.Sub(now) < (time.Duration(23*60+59) * time.Minute) {
		//	log.Println("jwt 续约")
		//	claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Hour * 24 * 60))
		//	tokenStr, err := token.SignedString([]byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"))
		//	if err != nil {
		//		// log
		//		log.Println("jwt 续约失败：", err)
		//	}
		//
		//	ctx.Header("x-jwt-token", tokenStr)
		//}

		ctx.Set("claims", claims)
	}
}
