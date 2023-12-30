package repository

import (
	"context"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(db *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{dao: db, cache: c}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	// 缓存里面有数据
	// 缓存里面没有数据
	// 缓存出错了，你也不知道有没有数据
	if err != nil {
		return domain.User{}, err
	}
	// 没有数据
	//if errors.Is(err, cache.ErrKeyNotExist) {
	//	// 去数据库里面找
	//}

	// 这里怎么办？ 比如 err = io.EOF
	// 要不要去数据库加载
	// redis出现缓存穿透、雪崩等redis崩掉了，那要不要去数据库里面找？会不会将数据库弄蹦？
	// 不去，那万一io.EOF是偶发性的呢？

	// 选加载，需要做好兜底，万一 Redis 真的蹦了，要保护你的数据库
	// 1. 数据库限流 - ORM的 middleware,但不能用redis来做限流，因redis已经崩掉了，用内存做单机限流
	// 选不加载，用户体验差一点
	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = domain.User{
		Id:       ue.Id,
		Email:    ue.Email,
		Password: ue.Password,
	}
	err = r.cache.Set(ctx, u)
	if err != nil {
		// 打日志，做监控
	}
	return u, err

	// 用缓存会面临的2个问题：
	// 1. 数据一致性问题
	// 2. 缓存蹦了（
	// 		1.数据库限流，
	//		2.使用二级缓存：（Redis+本地缓存）使用备用缓存（如本地缓存），Redis崩了，启用备用缓存）
	// 		3.（Redis + Redis）有一个高配置的Redis集群，还有一个廉价的低配Redis集群，高大上的蹦了，赶紧切换到低配的
}
