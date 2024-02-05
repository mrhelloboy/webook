package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

type JWTHandler struct {
	// access_token key
	atKey []byte
	// refresh_token key
	rtKey []byte
}

func newJWTHandler() JWTHandler {
	return JWTHandler{
		atKey: []byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"),
		rtKey: []byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORka"),
	}
}

func (h JWTHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	// 生成一个 JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.atKey)
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)

	fmt.Printf("-- token: %s\n", tokenStr)
	return nil
}

func (h JWTHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.rtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)

	fmt.Printf("--refresh token: %s\n", tokenStr)
	return nil
}

// ExtractToken 从请求头中获取 token
func ExtractToken(ctx *gin.Context) string {
	// 使用 jwt 校验
	tokenHeader := ctx.GetHeader("Authorization")
	// 从 tokenHeader 中获取 token
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}
	if segs[0] != "Bearer" || segs[1] == "" {
		return ""
	}
	return segs[1]
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid int64
}
