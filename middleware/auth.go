package middleware

import (
	"fmt"
	"strings"
	"tilapia/module/user"
	"tilapia/pkg/util"
	"tilapia/pkg/util/goslice"

	reqApi "tilapia/routes/api"

	"github.com/gin-gonic/gin"
	"github.com/infraboard/mcube/logger/zap"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 去除url中的/apiv1
		path := strings.TrimSpace(c.Request.URL.Path)
		if len(path) < 6 {
			zap.L().Error("request path is too short")
			util.JsonRespond(500, "request path is too short", "", c)
			return
		}
		path = path[6:]

		// 如果是登录操作请求 不检查Token
		if path == reqApi.LOGIN {
			return
		}

		token := c.Request.Header.Get("Authorization")
		fmt.Println(token)
		parts := strings.SplitN(token, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			util.JsonRespond(401, "API token required", "", c)
			c.Abort()
			return
		}
		token = parts[1]

		if token == "" {
			util.JsonRespond(401, "no login", "", c)
			c.Abort()
			return
		}

		login := &user.Login{
			Token: token,
		}
		if err := login.ValidateToken(); err != nil {
			util.JsonRespond(401, "Invalid API token, please login", "", c)
			zap.L().Error("校验token错误", zap.Error(err))
			c.Abort()
			return
		}

		//priv check
		role := &user.Role{
			ID: login.RoleId,
		}
		if err := role.Detail(); err != nil {
			util.JsonRespond(401, "Invalid API token, please login", "", c)
			return
		}
		c.Set("user_id", login.UserId)
		c.Set("username", login.Username)
		c.Set("email", login.Email)
		c.Set("truename", login.Truename)
		c.Set("mobile", login.Mobile)
		c.Set("role_name", role.Name)
		c.Set("privilege", role.Privilege)
		needLoginReq := []string{
			reqApi.LOGIN_STATUS,
			reqApi.LOGOUT,
			reqApi.MY_USER_SETTING,
			reqApi.MY_USER_PASSWORD,
		}
		if goslice.InSliceString(path, needLoginReq) {
			c.Next()
			return
		}
		if pass := user.CheckHavePriv(path, role.Privilege); !pass {
			util.JsonRespond(401, "no priv", "", c)
			return
		}

		c.Next()
	}
}
