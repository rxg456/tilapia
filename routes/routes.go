package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tilapia/controller"
	"tilapia/logger"
	"tilapia/middleware"
	reqApi "tilapia/routes/api"
	"tilapia/settings"
)

func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, settings.Conf.Version)
	})
	apiv1 := r.Group("/apiv1", middleware.TokenAuthMiddleware())
	// apiv1 := r.Group("/apiv1")
	{
		// 用户
		apiv1.POST(reqApi.LOGIN, controller.Login)
		apiv1.GET(reqApi.LOGIN_STATUS, controller.LoginStatus)
		apiv1.POST(reqApi.LOGOUT, controller.Logout)

		apiv1.POST(reqApi.MY_USER_SETTING, controller.MyUserSetting)
		apiv1.POST(reqApi.MY_USER_PASSWORD, controller.MyUserPassword)

		apiv1.POST(reqApi.USER_ADD, controller.UserAdd)
		apiv1.POST(reqApi.USER_UPDATE, controller.UserUpdate)
		apiv1.GET(reqApi.USER_LIST, controller.UserList)
		apiv1.GET(reqApi.USER_EXISTS, controller.UserExists)
		apiv1.GET(reqApi.USER_DETAIL, controller.UserDetail)
		apiv1.POST(reqApi.USER_DELETE, controller.UserDelete)

		apiv1.GET(reqApi.USER_ROLE_PRIV_LIST, controller.RolePrivList)
		apiv1.POST(reqApi.USER_ROLE_ADD, controller.RoleAdd)
		apiv1.POST(reqApi.USER_ROLE_UPDATE, controller.RoleUpdate)
		apiv1.GET(reqApi.USER_ROLE_LIST, controller.RoleList)
		apiv1.GET(reqApi.USER_ROLE_DETAIL, controller.RoleDetail)
		apiv1.POST(reqApi.USER_ROLE_DELETE, controller.RoleDelete)
	}

	return r
}
