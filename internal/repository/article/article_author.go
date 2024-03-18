package article

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/pkg/logger"

	"github.com/mrhelloboy/wehook/internal/repository"

	"github.com/ecodeclub/ekit/slice"

	"github.com/mrhelloboy/wehook/internal/domain"
	daoArt "github.com/mrhelloboy/wehook/internal/repository/dao/article"
)

// AuthorRepository 制作库接口
type AuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type cachedAuthorRepo struct {
	dao      daoArt.AuthorDAO
	userRepo repository.UserRepository
	cache    cache.ArticleCache
	l        logger.Logger
}

func NewCachedAuthorRepo(dao daoArt.AuthorDAO, userRepo repository.UserRepository, c cache.ArticleCache, l logger.Logger) AuthorRepository {
	return &cachedAuthorRepo{
		dao:      dao,
		userRepo: userRepo,
		cache:    c,
		l:        l,
	}
}

func (c *cachedAuthorRepo) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		_ = c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.UpdateById(ctx, daoArt.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *cachedAuthorRepo) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		_ = c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.Insert(ctx, daoArt.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *cachedAuthorRepo) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	// 缓存
	if err == nil {
		// 删除旧缓存（第一页数据已经增多，所以删除缓存）
		_ = c.cache.DelFirstPage(ctx, art.Author.Id)
		err := c.cache.Set(ctx, art)
		if err != nil {
			c.l.Warn("同步文章时，缓存失败", logger.Error(err))
		}
	}
	return id, err
}

func (c *cachedAuthorRepo) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (c *cachedAuthorRepo) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit <= 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		// 命中缓存
		if err == nil {
			// 预缓存第一条数据内容
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[daoArt.Article, domain.Article](res, func(idx int, src daoArt.Article) domain.Article {
		return c.toDomain(src)
	})

	// 缓存第一页数据
	if offset == 0 && limit <= 100 {
		go func() {
			err := c.cache.SetFirstPage(ctx, uid, data)
			if err != nil {
				c.l.Error("回写缓存失败", logger.Error(err))
			}
			c.preCache(ctx, data)
		}()
	}

	return data, nil
}

func (c *cachedAuthorRepo) GetById(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(data), nil
}

func (c *cachedAuthorRepo) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 获取作者信息
	author, err := c.userRepo.FindById(ctx, art.AuthorId)
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id:   author.Id,
			Name: author.Nickname,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (c *cachedAuthorRepo) toEntity(article domain.Article) daoArt.Article {
	return daoArt.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToUint8(),
	}
}

func (c *cachedAuthorRepo) toDomain(art daoArt.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
}

func (c *cachedAuthorRepo) preCache(ctx context.Context, data []domain.Article) {
	// 预缓存，且只缓存第一条数据
	// 这里只预加载长度小于1M的数据
	// 因为缓存数据是存放在内存中的，如果数据过大会占用过多内存
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := c.cache.Set(ctx, data[0])
		if err != nil {
			c.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}
