package controller

import (
	"tilapia/module/user"
	"tilapia/pkg/util"

	"github.com/gin-gonic/gin"
)

type UserSettingForm struct {
	Truename string `form:"truename"`
	Mobile   string `form:"mobile"`
}

type UserPasswordForm struct {
	Password    string `form:"password"`
	NewPassword string `form:"new_password"`
}

func MyUserSetting(c *gin.Context) {
	var form UserSettingForm
	if err := c.ShouldBind(&form); err != nil {
		util.JsonRespond(500, "Failure to parse", "", c)
		return
	}

	u := &user.User{
		ID:       c.GetInt("user_id"),
		Truename: form.Truename,
		Mobile:   form.Mobile,
	}
	if err := u.UserSettingUpdate(); err != nil {
		util.JsonRespond(500, "Failure to userSetting", "", c)
		return
	}
	util.JsonRespond(200, "UserSetting success", "", c)
}

func MyUserPassword(c *gin.Context) {
	var form UserPasswordForm
	if err := c.ShouldBind(&form); err != nil {
		util.JsonRespond(500, "Failure to parse", "", c)
		return
	}
	user := &user.User{
		ID: c.GetInt("user_id"),
	}
	if err := user.Detail(); err != nil {
		util.JsonRespond(500, "user detail error", "", c)
		return
	}

	err := util.CheckPasswordHash(form.Password, user.Password)
	if !err {
		util.JsonRespond(500, "user detail error", "", c)
		return
	}

	user.Password = form.NewPassword
	if err := user.UpdatePassword(); err != nil {
		util.JsonRespond(500, "controller.MyUserPassword error", err, c)
		return
	}

	util.JsonRespond(200, "controller.MyUserPassword success", "", c)
}
