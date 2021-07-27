package routes

import (
	"net/http"
	"tilapia/controller"
	"tilapia/logger"
	"tilapia/middleware"
	"tilapia/settings"

	"github.com/gin-gonic/gin"
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
	{
		// 用户登录
		apiv1.POST("/user/login", controller.Login)
		apiv1.POST("/user/logout", controller.Logout)
		apiv1.GET("/user/perms/:id", controller.GetUserMenu)

		apiv1.GET("/user", controller.GetUsers)
		apiv1.POST("/user", controller.PostUser)
		apiv1.PUT("/user/:id", controller.PutUser)
		apiv1.PATCH("/user/:id", controller.PatchUser)
		apiv1.DELETE("/user/:id", controller.DeleteUser)

		apiv1.GET("/perms", controller.GetPerms)
		apiv1.POST("/perms", controller.PostPerms)
		apiv1.PUT("/perms/:id", controller.PutPerms)
		apiv1.DELETE("/perms/:id", controller.DeletePerms)
		apiv1.GET("/perms/lists", controller.GetAllPerms)

		apiv1.GET("/roles", controller.GetRole)
		apiv1.POST("/roles", controller.PostRole)
		apiv1.PUT("/roles/:id", controller.PutRole)
		apiv1.DELETE("/roles/:id", controller.DeleteRole)
		apiv1.GET("/roles/:id/permissions", controller.GetRolePerms)
		apiv1.POST("/roles/:id/permissions", controller.PostRolePerms)

	}

	return r
}
