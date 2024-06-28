package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/pkg/ginx"
	"github.com/mrhelloboy/wehook/pkg/migrator/validator"

	"github.com/mrhelloboy/wehook/pkg/gormx/connpool"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/mrhelloboy/wehook/pkg/migrator"
	"github.com/mrhelloboy/wehook/pkg/migrator/events"
	"gorm.io/gorm"
)

// Scheduler 用来统一管理整个迁移过程
// 非必须
type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.Logger
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer
	fulls      map[string]func()
}

func NewScheduler[T migrator.Entity](src *gorm.DB, dst *gorm.DB, l logger.Logger, pool *connpool.DoubleWritePool, producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		src:     src,
		dst:     dst,
		pool:    pool,
		l:       l,
		pattern: connpool.PatternSrcOnly,
		cancelFull: func() {
		},
		cancelIncr: func() {
		},
		producer: producer,
	}
}

// 这一个也不是必须的，就是你可以考虑利用配置中心，监听配置中心的变化
// 把全量校验，增量校验做成分布式任务，利用分布式任务调度平台来调度
func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	// 将这个暴露为 HTTP 接口
	// 你可以配上对应的 UI
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))
	server.POST("/src_first", ginx.Wrap(s.SrcFirst))
	server.POST("/dst_first", ginx.Wrap(s.DstFirst))
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrementValidation))
	server.POST("/incr/start", ginx.WrapBody[StartIncrRequest](s.StartIncrementValidation))
}

// ---- 下面是四个阶段 ---- //

// SrcOnly 只读写源表
func (s *Scheduler[T]) SrcOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.UpdatePattern(connpool.PatternSrcOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context,
	req StartIncrRequest,
) (ginx.Result, error) {
	// 开启增量校验
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, nil
	}
	v.Incr().Utime(req.Utime).
		SleepInterval(time.Duration(req.Interval) * time.Millisecond)

	go func() {
		var ctx context.Context
		ctx, s.cancelIncr = context.WithCancel(context.Background())
		cancel()
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
	return ginx.Result{
		Msg: "启动增量校验成功",
	}, nil
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// StartFullValidation 全量校验
func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	// 可以考虑去重的问题
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		// 先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type StartIncrRequest struct {
	Utime int64 `json:"utime"`
	// 毫秒数
	// json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}
