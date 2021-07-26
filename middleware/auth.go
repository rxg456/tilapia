package middleware

import (
	"tilapia/dao/mysql"
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
