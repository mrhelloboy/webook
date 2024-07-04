package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"go.uber.org/atomic"

	"github.com/gin-gonic/gin"
)

// Builder 将每次请求及响应信息都记录到日志中
// 注意：
// 1. 小心日志内容过多。URL 可能很长，请求体，响应体都可能很大，需要考虑是否完全输出到日志里面的问题
// 2. 用户可能更换不同的日志框架，所以需要足够的灵活性
// 3. 考虑动态开关，结合监听配置文件，及小心并发安全
type Builder struct {
	allowReqBody  *atomic.Bool
	allowRespBody *atomic.Bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *Builder {
	return &Builder{
		loggerFunc:    fn,
		allowReqBody:  atomic.NewBool(false),
		allowRespBody: atomic.NewBool(false),
	}
}

func (b *Builder) AllowReqBody(val bool) *Builder {
	b.allowReqBody.Store(val)
	return b
}

func (b *Builder) AllowRespBody(val bool) *Builder {
	b.allowRespBody.Store(val)
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 借助 http header 来传递超时信息
		// timeout := ctx.GetHeader("x-timeout").(int64)
		// t := time.UnixMilli(timeout)
		// reqCtx, cancel := context.WithDeadline(ctx, t)
		start := time.Now()
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			Path:   url,
		}
		if b.allowReqBody.Load() && ctx.Request.Body != nil {
			body, _ := ctx.GetRawData() // io.ReadAll(ctx.Request.Body)
			// 因为 ctx.Request.Body 是一个流，io.ReadAll 会将其读取完毕，所以需要放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			if len(body) > 1024 {
				body = body[:1024]
			}
			// 该操作是一个很消耗 CPU 和内存的操作，因为会引起复制
			al.ReqBody = string(body)
		}

		if b.allowRespBody.Load() && ctx.Writer != nil {
			w := &responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
			ctx.Writer = w
		}

		defer func() {
			al.Duration = time.Since(start).String()
			b.loggerFunc(ctx, al)
		}()

		ctx.Next()
	}
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.al.RespBody = s
	return w.ResponseWriter.WriteString(s)
}

type AccessLog struct {
	Method   string // HTTP 请求方法
	Path     string // 请求的 URL
	Duration string
	ReqBody  string
	RespBody string
	Status   int
}
