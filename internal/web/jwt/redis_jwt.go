package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	// AtKey access_token key
	AtKey = []byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORkW")
	// RtKey refresh_token key
	RtKey = []byte("Xorxo9JJUq0v0PbqVbrRjThJXTCGORka")
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}

func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)

	fmt.Printf("--refresh token: %s\n", tokenStr)
	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	// 将 token 和 refresh token 设置为空
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	uc := ctx.MustGet("claims").(*UserClaims)
	err := h.cmd.Set(ctx, fmt.Sprintf("user:ssid:%s", uc.Ssid), "", time.Hour*24*7).Err()
	return err
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.cmd.Exists(ctx, fmt.Sprintf("user:ssid:%s", ssid)).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil
	case err == nil:
		if cnt == 0 {
			return nil
		}
		return errors.New("session 已经无效，用于已退出")
	default:
		return err
	}
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
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

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	// 生成一个 JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)

	fmt.Printf("-- token: %s\n", tokenStr)
	return nil
}
