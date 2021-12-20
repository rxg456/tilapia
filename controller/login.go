package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tilapia/module/user"
	"tilapia/pkg/util"
	"tilapia/pkg/util/gois"
)

type LoginBind struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func Logout(c *gin.Context) {
	login := &user.Login{
		UserId: c.GetInt("user_id"),
	}
	if err := login.Logout(); err != nil {
		zap.L().Error("controller.Logout error", zap.Error(err))
		util.JsonRespond(500, "Logout error", "", c)
		return
	}
	util.JsonRespond(200, "logout success", "", c)
}

func LoginStatus(c *gin.Context) {
	privilege, _ := c.Get("privilege")

	data := make(map[string]interface{})
	data["is_login"] = 1
	data["user_id"] = c.GetInt("user_id")
	data["username"] = c.GetString("username")
	data["email"] = c.GetString("email")
	data["truename"] = c.GetString("truename")
	data["mobile"] = c.GetString("mobile")
	data["role_name"] = c.GetString("role_name")
	data["privilege"] = privilege

	util.JsonRespond(200, "", data, c)
}

func Login(c *gin.Context) {
	var dataResource LoginBind
	if err := c.BindJSON(&dataResource); err != nil {
		util.JsonRespond(500, "Login with invalid param", "", c)
		zap.L().Error("controller.Login with invalid param:", zap.Error(err))
		return
	}

	login := &user.Login{
		Password: dataResource.Password,
	}

	if gois.IsEmail(dataResource.Username) {
		login.Email = dataResource.Username
	} else {
		login.Username = dataResource.Username
	}

	if err := login.Login(); err != "" {
		util.JsonRespond(500, err, "", c)
		return
	}

	userInfo := map[string]interface{}{
		"token": login.Token,
	}

	util.JsonRespond(200, "", userInfo, c)
	return
}
