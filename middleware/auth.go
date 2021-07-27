package middleware

import (
	"fmt"
	"tilapia/dao/mysql"
	"tilapia/dao/redis"
	"tilapia/models"
	"tilapia/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		// 如果是登录操作请求 不检查Token
		uri := c.Request.URL.String()
		if uri == "/apiv1/user/login" {
			return
		}

		token := c.Request.Header.Get("Authorization")

		if token == "" {
			util.JsonRespond(401, "API token required", "", c)
			c.Abort()
			return
		}

		e := mysql.DB.Where("access_token = ?", token).First(&user).Error
		if e != nil {
			util.JsonRespond(401, "Invalid API token, please login", "", c)
			c.Abort()
		}

		if user.IsActive == 1 && user.TokenExpired > time.Now().Unix() {
			user.TokenExpired = time.Now().Unix() + 86400
			mysql.DB.Save(&user)
		}

		c.Set("UserIsSupper", user.IsSupper)
		c.Set("UserRid", user.Rid)
		c.Set("Uid", user.ID)
		c.Next()
	}
}

func PermissionCheckMiddleware(c *gin.Context, perm string) bool {
	UserIsSupper, _ := c.Get("UserIsSupper")

	// 超级用户不做权限检查
	if UserIsSupper != 1 {
		key := redis.RoleRermsSetKey
		UserRid, _ := c.Get("UserRid")
		str := fmt.Sprintf("%v", UserRid)
		user_key := key + str

		// redis是否有key值
		if err := redis.Rdb.Exists(user_key).Val(); err != 1 {
			rid, _ := UserRid.(int)
			models.SetRolePermToSet(user_key, rid)
		}

		// 检查对应的set是否有该角色权限
		return redis.CheckMemberByKey(user_key, perm)
	} else {
		return true
	}
}
