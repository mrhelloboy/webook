package repository

import (
	"context"
	"database/sql"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(db dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{dao: db, cache: c}
}

func (r *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.Get(ctx, id)
	// 缓存里面有数据
	// 缓存里面没有数据
	// 缓存出错了，你也不知道有没有数据
	if err == nil {
		return u, nil
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

	u = r.entityToDomain(ue)

	// 可以异步也可以同步
	go func() {
		_ = r.cache.Set(ctx, u)
		//if err != nil {
		//	// 打日志，做监控
		//}
	}()

	return u, err

	// 用缓存会面临的2个问题：
	// 1. 数据一致性问题
	// 2. 缓存蹦了（
	// 		1.数据库限流，
	//		2.使用二级缓存：（Redis+本地缓存）使用备用缓存（如本地缓存），Redis崩了，启用备用缓存）
	// 		3.（Redis + Redis）有一个高配置的Redis集群，还有一个廉价的低配Redis集群，高大上的蹦了，赶紧切换到低配的
}

func (r *CachedUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	user, err := r.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id:            u.Id,
		Email:         sql.NullString{String: u.Email, Valid: u.Email != ""},
		Password:      u.Password,
		Nickname:      u.Nickname,
		Phone:         sql.NullString{String: u.Phone, Valid: u.Phone != ""},
		WechatOpenId:  sql.NullString{String: u.WechatInfo.OpenID, Valid: u.WechatInfo.OpenID != ""},
		WechatUnionId: sql.NullString{String: u.WechatInfo.UnionID, Valid: u.WechatInfo.UnionID != ""},
		Ctime:         u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Nickname: u.Nickname,
		WechatInfo: domain.WechatInfo{
			OpenID:  u.WechatOpenId.String,
			UnionID: u.WechatUnionId.String,
		},
		Ctime: time.UnixMilli(u.Ctime),
	}
}
