package controller

import (
	"tilapia/dao/mysql"
	"tilapia/dao/redis"
	"tilapia/models"
	"tilapia/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type LoginResource struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Secret   string `form:"secret"`
}

func Login(c *gin.Context) {
	var dataResource LoginResource
	var user models.User
	var expiration = time.Duration(86400) * time.Second

	if err := c.BindJSON(&dataResource); err != nil {
		util.JsonRespond(500, "Login with invalid param", "", c)
		zap.L().Error("Login with invalid param", zap.Error(err))
		return
	}

	username := dataResource.Username
	password := dataResource.Password
	key := username + "_login"
	if err := mysql.DB.Where("name = ?", username).First(&user).Error; err != nil {
		util.JsonRespond(500, "用户不存在！", "", c)
		zap.L().Error("用户不存在:", zap.Error(err))
		return
	}
	if user.IsActive == 1 {
		err := util.CheckPasswordHash(password, user.PasswordHash)
		if !err {
			// 记录用户验证失败次数
			// 检查key是否存在 1: 存在， 0: 不存在
			if redis.Rdb.Exists(key).Val() == 1 {
				// 获取key值
				num, _ := redis.Rdb.Get(key).Int()
				// 验证超过3次，将锁定用户
				if num > 3 {
					util.JsonRespond(401, "用户已禁用,请联系管理员", "", c)
					return
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
			util.JsonRespond(401, "用户名或密码错误，连续3次错误将会被禁用", "", c)
			return
		} else {
			// 生成token
			token := uuid.New().String()
			user.AccessToken = token
			user.TokenExpired = time.Now().Unix() + 86400

			// 提交更改
			mysql.DB.Save(&user)

			// 获取用户的权限列表
			var permissions []string
			if user.IsSupper != 1 {
				permissions = user.ReturnPermissions()
			}

			data := make(map[string]interface{})
			data["rid"] = user.Rid
			data["token"] = token
			data["is_supper"] = user.IsSupper
			data["nickname"] = user.Nickname
			data["permissions"] = permissions

			// 登录成功
			if err := redis.SetValByKey(key, 0, expiration); err != nil {
				zap.L().Error("login success set redis faild", zap.Error(err))
			}

			util.JsonRespond(200, "", data, c)
			return

		}
	} else {
		util.JsonRespond(500, "用户被禁用，请联系管理员！", "", c)
		return
	}

}
