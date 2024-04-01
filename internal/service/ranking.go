package service

import (
	"context"
	"math"
	"time"

	"github.com/mrhelloboy/wehook/internal/repository"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingSrv struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	scoreFunc func(t time.Time, likeCnt int64) float64 // 不能返回负数
	load      int64                                    // 负载
}

func NewBatchRankingSrv(artSvc ArticleService, intrSvc InteractiveService, repo repository.RankingRepository) RankingService {
	return &BatchRankingSrv{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			// 假设 likeCnt 是点赞数，t 是文章发布时间
			// 可以根据需要定义自己的评分函数
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec+2, 1.5)
		},
	}
}

func (s *BatchRankingSrv) TopN(ctx context.Context) error {
	arts, err := s.topN(ctx)
	if err != nil {
		return err
	}
	return s.repo.ReplaceTopN(ctx, arts)
}

func (s *BatchRankingSrv) topN(ctx context.Context) ([]domain.Article, error) {
	// 只取 7 天内的数据
	now := time.Now()
	// 先拿一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	// 优先队列
	topN := queue.NewConcurrentPriorityQueue[Score](s.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})

	for {
		// 这里拿了一批
		arts, err := s.artSvc.ListPub(ctx, now, offset, s.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		// 找对应的点赞数据
		intrs, err := s.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := intrs[art.Id]
			score := s.scoreFunc(art.Utime, intr.LikeCnt)
			// 考虑这个 score 在不在前 100 名内
			// 拿到热度最低的
			err = topN.Enqueue(Score{art: art, score: score})
			// 这种写法，要求 topN 已经满了
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					_ = topN.Enqueue(Score{art: art, score: score})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}

		// 处理完一批数据，要不要进入下一批？
		if len(arts) < s.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 这一批都没有取够，当前肯定没有下一批了
			// 或者已经取到了 7 天之前的数据了，说明可以中断了
			break
		}
		offset = offset + len(arts)
	}

	// 最后得出结果
	res := make([]domain.Article, s.n)
	for i := s.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 说明取完了，不够 n
			break
		}
		res[i] = val.art
	}
	return res, nil
}
