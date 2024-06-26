package service

import (
	"context"
	"testing"
	"time"

	intrv1 "github.com/mrhelloboy/wehook/api/proto/gen/intr/v1"

	domain2 "github.com/mrhelloboy/wehook/interactive/domain"

	"github.com/mrhelloboy/wehook/internal/domain"
	svcmocks "github.com/mrhelloboy/wehook/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBatchRankingSrv_TopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (ArticleService, intrv1.InteractiveServiceClient)
		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			mock: func(ctrl *gomock.Controller) (ArticleService, intrv1.InteractiveServiceClient) {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).Return([]domain.Article{
					{Id: 1, Utime: now, Ctime: now},
					{Id: 2, Utime: now, Ctime: now},
					{Id: 3, Utime: now, Ctime: now},
				}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).Return([]domain.Article{}, nil)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).Return(map[int64]domain2.Interactive{
					1: {BizId: 1, LikeCnt: 1},
					2: {BizId: 2, LikeCnt: 2},
					3: {BizId: 3, LikeCnt: 3},
				}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).Return(map[int64]domain2.Interactive{}, nil)
				return artSvc, intrSvc
			},
			wantArts: []domain.Article{
				{Id: 3, Utime: now, Ctime: now},
				{Id: 2, Utime: now, Ctime: now},
				{Id: 1, Utime: now, Ctime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc := tc.mock(ctrl)
			svc := NewBatchRankingSrv(artSvc, intrSvc, nil).(*BatchRankingSrv)
			// 为了方便测试，修改batchSize，n 和 scoreFunc
			svc.batchSize = 3
			svc.n = 3
			svc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
