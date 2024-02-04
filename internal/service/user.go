package service

import (
	"context"
	"errors"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService interface {
	Login(ctx context.Context, email, password string) (domain.User, error)
	Signup(ctx context.Context, u domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
}

type UserSvc struct {
	repo repository.UserRepository
}

func NewUserSvc(repo repository.UserRepository) UserService {
	return &UserSvc{
		repo: repo,
	}
}

func (svc *UserSvc) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先查询用户是否存在
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// 查询时出错，比如超时
	if err != nil {
		return domain.User{}, err
	}
	// 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserSvc) Signup(ctx context.Context, u domain.User) error {
	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserSvc) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 快路径
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 业务中大多数请求时进入这个逻辑
		// err 为 nil，会进入该逻辑，说明用户存在
		return u, err
	}
	// 慢路径
	// 在系统资源不足，触发降级之后，不执行慢路径
	//if ctx.Value("downgrade") == "true" {
	//	return domain.User{}, errors.New("系统降级了")
	//}
	err = svc.repo.Create(ctx, domain.User{Phone: phone})
	// 注册有问题，但是又不是用户手机号码冲突，说明系统异常
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return domain.User{}, err
	}

	// todo: 这里有主从延迟的坑
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *UserSvc) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	// 快路径
	u, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 业务中大多数请求时进入这个逻辑
		// err 为 nil，会进入该逻辑，说明用户存在
		return u, err
	}

	err = svc.repo.Create(ctx, domain.User{WechatInfo: info})
	// 注册有问题，但是又不是用户手机号码冲突，说明系统异常
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return domain.User{}, err
	}

	// todo: 这里有主从延迟的坑
	return svc.repo.FindByWechat(ctx, info.UnionID)
}

func (svc *UserSvc) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}
