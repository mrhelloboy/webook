package service

import (
	"context"
	"errors"
	"testing"

	repomocks "github.com/mrhelloboy/wehook/internal/repository/article/mocks"

	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/stretchr/testify/assert"

	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository/article"
	"go.uber.org/mock/gomock"
)

func Test_articleSvc_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (article.AuthorRepository, article.ReaderRepository)
		art     domain.Article
		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (article.AuthorRepository, article.ReaderRepository) {
				author := repomocks.NewMockAuthorRepository(ctrl)
				reader := repomocks.NewMockReaderRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1, // 使用制作库的 ID
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return author, reader
			},
			art: domain.Article{
				Title:   "test",
				Content: "test",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: nil,
			wantId:  1,
		},
		{
			name: "修改并发表成功",
			mock: func(ctrl *gomock.Controller) (article.AuthorRepository, article.ReaderRepository) {
				author := repomocks.NewMockAuthorRepository(ctrl)
				reader := repomocks.NewMockReaderRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2, // 使用制作库的 ID
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(2), nil)
				return author, reader
			},
			art: domain.Article{
				Id:      2,
				Title:   "test",
				Content: "test",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: nil,
			wantId:  2,
		},
		{
			name: "保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (article.AuthorRepository, article.ReaderRepository) {
				author := repomocks.NewMockAuthorRepository(ctrl)
				reader := repomocks.NewMockReaderRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock db error"))
				return author, reader
			},
			art: domain.Article{
				Title:   "test",
				Content: "test",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: errors.New("mock db error"),
			wantId:  0,
		},
		{
			name: "保存到制作库成功，重试到线上库成功",
			mock: func(ctrl *gomock.Controller) (article.AuthorRepository, article.ReaderRepository) {
				author := repomocks.NewMockAuthorRepository(ctrl)
				reader := repomocks.NewMockReaderRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1, // 使用制作库的 ID
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock db error"))
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1, // 使用制作库的 ID
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return author, reader
			},
			art: domain.Article{
				Title:   "test",
				Content: "test",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: nil,
			wantId:  1,
		},
		{
			name: "保存到制作库成功，重试全部失败",
			mock: func(ctrl *gomock.Controller) (article.AuthorRepository, article.ReaderRepository) {
				author := repomocks.NewMockAuthorRepository(ctrl)
				reader := repomocks.NewMockReaderRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1, // 使用制作库的 ID
					Title:   "test",
					Content: "test",
					Author: domain.Author{
						Id: 123,
					},
				}).Times(3).Return(int64(0), errors.New("mock db error"))
				return author, reader
			},
			art: domain.Article{
				Title:   "test",
				Content: "test",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: errors.New("mock db error"),
			wantId:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			author, reader := tc.mock(ctrl)
			svc := NewArticleSvc(author, reader, &logger.NopLogger{})
			id, err := svc.Publish(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
