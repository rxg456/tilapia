package middleware

import (
	"fmt"
	"strings"
	"tilapia/module/user"
	"tilapia/pkg/util"
	"tilapia/pkg/util/goslice"

	reqApi "tilapia/routes/api"

	"github.com/gin-gonic/gin"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是登录操作请求 不检查Token
		path := c.Request.URL.String()
		if path == "/apiv1/login" {
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

		fmt.Println(token)
		if token == "" {
			util.JsonRespond(401, "no login", "", c)
			c.Abort()
			return
		}

		login := &user.Login{
			Token: token,
		}

		if err := login.ValidateToken(token); err != nil {
			util.JsonRespond(401, "Invalid API token, please login", "", c)
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

// func PermissionCheckMiddleware(c *gin.Context, perm string) bool {
// 	UserIsSupper, _ := c.Get("UserIsSupper")

// 	// 超级用户不做权限检查
// 	if UserIsSupper != 1 {
// 		key := redis.RoleRermsSetKey
// 		UserRid, _ := c.Get("UserRid")
// 		str := fmt.Sprintf("%v", UserRid)
// 		user_key := key + str

// 		// redis是否有key值
// 		if err := redis.Rdb.Exists(user_key).Val(); err != 1 {
// 			rid, _ := UserRid.(int)
// 			models.SetRolePermToSet(user_key, rid)
// 		}

// 		// 检查对应的set是否有该角色权限
// 		return redis.CheckMemberByKey(user_key, perm)
// 	} else {
// 		return true
// 	}
// }
