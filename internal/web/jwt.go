package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTHandler struct {
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
	tokenStr, err := token.SignedString([]byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW"))
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)

	fmt.Printf("-- token: %s\n", tokenStr)
	return nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
