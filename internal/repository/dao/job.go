package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

// Preempt 抢占任务
func (g *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	// 高并发情况下，大部分都是陪太子读书
	// 100 个 goroutine
	// 要转几次？所有 goroutine 执行的循环次数加在一起是 1+2+3+4+5+...+99+100
	// 特定一个 goroutine，最差情况下，要循环一百次
	db := g.db.WithContext(ctx)
	for {
		now := time.Now()
		var j Job
		// 分布式任务调度系统
		// 1. 一次拉一批，我一次性取出 100 条来，然后，我随机从某一条开始，向后开始抢占
		// 2. 我搞一个随机偏移量，0-100 生成一个随机偏移量。兜底：第一轮没查到，偏移量回归到 0
		// 3. 我搞一个 id 取余分配，status = ？AND next_time <= ? AND id%10 = ? 兜底：不加余数条件，取 next_time 最老的
		err := db.WithContext(ctx).Where("status = ? AND next_time <= ?", jobStatusWaiting, now).First(&j).Error
		// 抢占任务
		if err != nil {
			// 没有任务
			return Job{}, err
		}
		// 两个 goroutine 都拿到 id=1 的数据情况
		// 使用乐观锁，CAS（compare AND swap）操作
		// 面试亮点：使用乐观锁取代 FOR UPDATE
		// 面试套路（性能优化）：增将用了 FOR UPDATE => 性能差，还会有死锁 => 优化成了乐观锁
		res := db.Where("id = ? AND version = ?", j.Id, j.Version).Model(&Job{}).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"utime":   now,
				"version": j.Version + 1,
			})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			// 抢占失败，只能继续下一轮
			continue
		}
		return j, nil
	}
}

// Release 释放任务
func (g *GORMJobDAO) Release(ctx context.Context, id int64) error {
	// 问题：要不要检测 status 或者 version？
	// WHERE version = ？
	// 这是需要的。
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

// UpdateUtime 更新 utime
func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"utime": time.Now().UnixMilli(),
	}).Error
}

// UpdateNextTime 更新 next_time
func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"next_time": next.UnixMilli(),
	}).Error
}

// Stop 停止任务
func (g *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"status": jobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

type Job struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Cfg      string
	Executor string
	Name     string `gorm:"unique"`
	// 问题：那些任务可以抢？那些任务已经被人占着？那些任务永远不会被运行
	// 用标记标记
	Status int
	// 问题：定时任务，我怎么知道，已经到时间了呢？
	// NextTime 下一次被调度的时间
	// next_time <= now 这样一个查询条件
	// and status = 0
	// 要建立索引 （在 next_time 和 status 的联合索引）
	NextTime int64 `gorm:"index"`
	// cron 表达式
	Cron    string
	Version int
	Ctime   int64
	Utime   int64
}

const (
	jobStatusWaiting int = iota
	jobStatusRunning
	jobStatusPaused // 暂停调度
)
