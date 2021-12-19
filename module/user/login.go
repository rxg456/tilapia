package user

import (
	"errors"
	"time"

	"tilapia/dao/redis"
	"tilapia/pkg/util"

	"github.com/google/uuid"
	"github.com/infraboard/mcube/logger/zap"
)

type Login struct {
	UserId   int
	RoleId   int
	Username string
	Password string
	Email    string
	Truename string
	Mobile   string
	Token    string
}

func (login *Login) Logout() error {
    token := &Token{
        UserId: login.UserId,
    }
    return token.DeleteByUserId()
}

func (login *Login) Login() string {
	var expiration = time.Duration(86400) * time.Second
	var key string

	u := &User{}
	if login.Username != "" {
		u.Username = login.Username
		key = login.Username + "_login"
	}
	if login.Email != "" {
		u.Email = login.Email
		key = login.Email + "_login"
	}
	if err := u.Detail(); err != nil {
		return "用户或密码错误"
	}
	if u.Status != 1 {
		return "用户名不存在或未激活"
	}
	err := util.CheckPasswordHash(login.Password, u.Password)
	if !err {
		// 记录用户验证失败次数
		// 检查key是否存在 1: 存在， 0: 不存在
		if redis.Rdb.Exists(key).Val() == 1 {
			// 获取key值
			num, _ := redis.Rdb.Get(key).Int()
			// 验证超过3次，将锁定用户
			if num > 3 {
				return "用户已禁用,请联系管理员"
			}
			if err := redis.SetValByKey(key, num+1, expiration); err != nil {
				zap.L().Error("redis setvalbykey field", zap.Error(err))
			}
		} else {
			// 第一次登录失败
			if err := redis.SetValByKey(key, 1, expiration); err != nil {
				zap.L().Error("redis setvalbykey field", zap.Error(err))
			}
		}
		return "用户名或密码错误，连续3次错误将会被禁用"
	}
	login.UserId = u.ID
	if err := login.createToken(); err != nil {
		return "token create failed"
	}

	return ""
}

func (login *Login) createToken() error {
	utoken := uuid.New().String()
	login.Token = utoken

	token := &Token{
		UserId: login.UserId,
		Token:  login.Token,
		Expire: int(time.Now().Unix() + 86400*30),
	}
	if err := token.CreateOrUpdate(); err != nil {
		return err
	}
	return nil
}

func (login *Login) ValidateToken(t string) error {
	token := &Token{
		Token: t,
	}
	if ok := token.ValidateToken(); !ok {
		return errors.New("token check failed, maybe your account is logged in on another device or token expired")
	}

	//get user detail
	user := &User{
		ID: token.UserId,
	}
	if err := user.Detail(); err != nil {
		return errors.New("token check failed, user detail get failed")
	}

	if user.Status != 1 {
		return errors.New("user is locked")
	}

	login.UserId = user.ID
	login.Username = user.Username
	login.Email = user.Email
	login.Truename = user.Truename
	login.Mobile = user.Mobile
	login.RoleId = user.RoleId

	return nil
}
