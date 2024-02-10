package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

// 将每次请求及响应信息都记录到日志中

type Builder struct {
	allowReqBody  bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *Builder {
	return &Builder{
		loggerFunc: fn,
	}
}

func (b *Builder) AllowReqBody() *Builder {
	b.allowReqBody = true
	return b
}

func (b *Builder) AllowRespBody() *Builder {
	b.allowRespBody = true
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			Path:   url,
		}
		if b.allowReqBody && ctx.Request.Body != nil {
			body, _ := ctx.GetRawData() // io.ReadAll(ctx.Request.Body)
			// 因为 ctx.Request.Body 是一个流，io.ReadAll 会将其读取完毕，所以需要放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			if len(body) > 1024 {
				body = body[:1024]
			}
			// 该操作是一个很消耗 CPU 和内存的操作，因为会引起复制
			al.ReqBody = string(body)
		}

		if b.allowRespBody && ctx.Writer != nil {
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
