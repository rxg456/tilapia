package controller

import (
	"tilapia/module/user"
	"tilapia/pkg/util"
	"tilapia/pkg/util/gostring"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type QueryBind struct {
	Keyword string `form:"keyword"`
	Offset  int    `form:"offset"`
	Limit   int    `form:"limit" binding:"required,gte=1,lte=999"`
}

type UserForm struct {
	ID       int    `form:"id"`
	RoleId   int    `form:"role_id" binding:"required"`
	Username string `form:"username" binding:"required"`
	Password string `form:"password"`
	Email    string `form:"email" binding:"required"`
	Truename string `form:"truename"`
	Mobile   string `form:"mobile"`
	Status   int    `form:"status"`
}

type UserExistsQuery struct {
	ID       int    `form:"id"`
	Username string `form:"username"`
	Email    string `form:"email"`
}

func UserDetail(c *gin.Context) {
	id := gostring.Str2Int(c.Query("id"))
	if id == 0 {
		util.JsonRespond(500, "id cannot be empty", "", c)
		return
	}
	u := &user.User{
		ID: id,
	}
	if err := u.Detail(); err != nil {
		zap.L().Error("controller.UserDetail error", zap.Error(err))
		util.JsonRespond(500, "UserDetail error", "", c)
		return
	}
	detail := map[string]interface{}{
		"id":       u.ID,
		"role_id":  u.RoleId,
		"username": u.Username,
		"email":    u.Email,
		"truename": u.Truename,
		"mobile":   u.Mobile,
		"status":   u.Status,
	}
	util.JsonRespond(200, "UserDetail success", detail, c)
}

func UserDelete(c *gin.Context) {
	id := gostring.Str2Int(c.PostForm("id"))
	if id == 0 {
		zap.L().Error("id cannot be empty")
		util.JsonRespond(500, "id cannot be empty", "", c)
		return
	}
	u := &user.User{
		ID: id,
	}
	if err := u.Delete(); err != nil {
		zap.L().Error("delete user error", zap.Error(err))
		util.JsonRespond(500, "delete user error", "", c)
		return
	}
	util.JsonRespond(200, "delete user success", "", c)

}

func UserExists(c *gin.Context) {
	var query UserExistsQuery
	if err := c.ShouldBind(&query); err != nil {
		zap.L().Error("controller.UserExists with invalid param", zap.Error(err))
		util.JsonRespond(500, "UserExists with invalid param", "", c)
		return
	}
	u := &user.User{
		ID:       query.ID,
		Username: query.Username,
		Email:    query.Email,
	}
	exists, err := u.UserCheckExists()
	if err != nil {
		zap.L().Error("UserCheckExists error", zap.Error(err))
		util.JsonRespond(500, "UserCheckExists error", "", c)
		return
	}
	data := make(map[string]interface{})
	data["exists"] = exists

	util.JsonRespond(200, "userexists success", data, c)
}

func UserList(c *gin.Context) {
	var query QueryBind
	if err := c.ShouldBind(&query); err != nil {
		zap.L().Error("controller.UserList with invalid param", zap.Error(err))
		util.JsonRespond(500, "UserList with invalid param", "", c)
		return
	}
	u := &user.User{}
	list, err := u.List(query.Keyword, query.Offset, query.Limit)
	if err != nil {
		zap.L().Error("controller.UserList with invalid param", zap.Error(err))
		util.JsonRespond(500, "UserList with invalid param", "", c)
		return
	}
	total, err := u.Total(query.Keyword)
	if err != nil {
		zap.L().Error("userlist total error", zap.Error(err))
		util.JsonRespond(500, "userlist total error", "", c)
		return
	}

	var userList []map[string]interface{}
	for _, l := range list {
		userList = append(userList, map[string]interface{}{
			"id":              l.ID,
			"role_name":       l.RoleName,
			"username":        l.Username,
			"truename":        l.Truename,
			"email":           l.Email,
			"status":          l.Status,
			"last_login_time": l.LastLoginTime,
			"last_login_ip":   l.LastLoginIp,
		})
	}

	data := make(map[string]interface{})
	data["list"] = list
	data["total"] = total

	util.JsonRespond(200, "userlist success", data, c)
}

func UserUpdate(c *gin.Context) {
	var userForm UserForm
	if err := c.ShouldBind(&userForm); err != nil {
		zap.L().Error("controller.UserUpdate with invalid param", zap.Error(err))
		util.JsonRespond(500, "UserUpdate with invalid param", "", c)
		return
	}
	if userForm.ID == 0 {
		zap.L().Error("id cannot empty")
		util.JsonRespond(500, "id cannot empty", "", c)
		return
	}
	// if userForm.Password != "" && len(userForm.Password) != 32 {
	//     render.ParamError(c, "password param incorrect")
	//     return
	// }

	userCreateOrUpdate(c, userForm)
}

func UserAdd(c *gin.Context) {
	var userForm UserForm
	if err := c.ShouldBind(&userForm); err != nil {
		zap.L().Error("Invalid Add User Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Add User Data", "", c)
		return
	}

	// if len(userForm.Password) != 32 {
	// 	render.ParamError(c, "password param incorrect")
	// 	return
	// }

	userCreateOrUpdate(c, userForm)
}

func userCreateOrUpdate(c *gin.Context, userForm UserForm) {
	var (
		checkUsername, checkEmail *user.User
		exists                    bool
		err                       error
	)
	checkUsername = &user.User{
		ID:       userForm.ID,
		Username: userForm.Username,
	}
	exists, err = checkUsername.UserCheckExists()
	if err != nil {
		zap.L().Error("UserCheckExists error", zap.Error(err))
		util.JsonRespond(500, "UserCheckExists error", "", c)
		return
	}
	if exists {
		zap.L().Error("username have exists")
		util.JsonRespond(500, "username have exists", "", c)
		return
	}

	checkEmail = &user.User{
		ID:    userForm.ID,
		Email: userForm.Email,
	}
	exists, err = checkEmail.UserCheckExists()
	if err != nil {
		zap.L().Error("UserCheckExists Email error", zap.Error(err))
		util.JsonRespond(500, "UserCheckExists Email error", "", c)
		return
	}
	if exists {
		zap.L().Error("email have exists")
		util.JsonRespond(500, "email have exists", "", c)
		return
	}

	u := &user.User{
		ID:       userForm.ID,
		RoleId:   userForm.RoleId,
		Username: userForm.Username,
		Password: userForm.Password,
		Email:    userForm.Email,
		Truename: userForm.Truename,
		Mobile:   userForm.Mobile,
		Status:   userForm.Status,
	}
	if err := u.CreateOrUpdate(); err != nil {
		zap.L().Error("CreateOrUpdate user error", zap.Error(err))
		util.JsonRespond(500, "CreateOrUpdate user error", "", c)
		return
	}
	util.JsonRespond(200, "CreateOrUpdate user success", "", c)
}

// type LoginResource struct {
// 	Username string `form:"username"`
// 	Password string `form:"password"`
// }

// type UserResource struct {
// 	Name     string `form:"name"`
// 	Nickname string `form:"nickname"`
// 	Mobile   string `form:"mobile"`
// 	Email    string `form:"email"`
// 	Rid      int    `form:"rid"`
// 	Password string `form:"password"`
// 	IsActive int    `form:"isActive"`
// }

// // 用户菜单列表
// func GetUserMenu(c *gin.Context) {
// 	var perms []models.MenuPermissions
// 	var res util.SortMenuPerms
// 	var user models.User

// 	tmp := make(map[int]*models.MenuPerms)
// 	data := make(map[string]interface{})
// 	uid := c.Param("id")

// 	// 用户菜单列表
// 	key := redis.RoleMenuListKey
// 	str := fmt.Sprintf("%v", uid)
// 	key = key + str
// 	v, err := redis.Rdb.Get(key).Result()
// 	if err != nil {
// 		zap.L().Error("rolemenulist get faild", zap.Error(err))
// 	}
// 	if v != "" {
// 		data["lists"] = util.JsonUnmarshalFromString(v, &res)
// 		util.JsonRespond(200, "", data, c)
// 		return
// 	}

// 	// 获取用户角色id
// 	// uids := strings.SplitN(uid, "=", 2)
// 	// if err := mysql.DB.Where("id = ?", uids[1]).First(&user).Error; err != nil {
// 	// 	util.JsonRespond(500, "用户不存在！", "", c)
// 	// 	zap.L().Error("用户不存在:", zap.Error(err))
// 	// 	return
// 	// }
// 	if err := mysql.DB.Where("id = ?", uid).First(&user).Error; err != nil {
// 		util.JsonRespond(500, "用户不存在！", "", c)
// 		zap.L().Error("用户不存在:", zap.Error(err))
// 		return
// 	}
// 	// 超级用户返回所有菜单
// 	if user.Rid == 0 {
// 		mysql.DB.Model(&models.MenuPermissions{}).Where("type = 1").Order("id ASC").Find(&perms)
// 		for _, p := range perms {
// 			if p.Pid == 0 {
// 				tmp[p.ID] = &models.MenuPerms{
// 					ID:         p.ID,
// 					Pid:        p.Pid,
// 					Name:       p.Name,
// 					Type:       p.Type,
// 					Permission: p.Permission,
// 					Url:        p.Url,
// 					Icon:       p.Icon,
// 					Desc:       p.Desc,
// 				}
// 			} else {
// 				if x, ok := tmp[p.Pid]; ok {
// 					x.Children = append(x.Children, p)
// 				} else {
// 					tmp[p.Pid] = &models.MenuPerms{
// 						Children: []models.MenuPermissions{p},
// 					}
// 				}
// 			}
// 		}
// 	} else {
// 		// 普通用户 根据 rid 返回菜单项
// 		pids := []int{}
// 		mysql.DB.Table("menu_permissions").
// 			Select("menu_permissions.pid").
// 			Joins("left join role_permission_rels on menu_permissions.id = role_permission_rels.pid").
// 			Where("role_permission_rels.rid = ?", user.Rid).
// 			Pluck("DISTINCT menu_permissions.pid", &pids)

// 		mysql.DB.Model(&models.MenuPermissions{}).Where("type = 1").Find(&perms)
// 		for _, v := range perms {
// 			for _, p := range pids {
// 				if _, ok := tmp[v.ID]; !ok {
// 					tmp[v.ID] = &models.MenuPerms{
// 						ID:         v.ID,
// 						Pid:        v.Pid,
// 						Name:       v.Name,
// 						Type:       v.Type,
// 						Permission: v.Permission,
// 						Url:        v.Url,
// 						Icon:       v.Icon,
// 						Desc:       v.Desc,
// 					}
// 				}

// 				if p == v.ID {
// 					if x, ok := tmp[v.Pid]; ok {
// 						x.Children = append(x.Children, v)
// 					}
// 				}
// 			}
// 		}
// 	}

// 	for _, v := range tmp {
// 		if v.Pid == 0 {
// 			res = append(res, v)
// 		}
// 	}
// 	sort.Stable(res)

// 	redis.Rdb.Set(key, util.JSONMarshalToString(res), 0)

// 	data["lists"] = res
// 	util.JsonRespond(200, "", data, c)
// }

// // 用户列表
// func GetUsers(c *gin.Context) {
// 	var res []models.User
// 	var count int64

// 	name := c.Query("name")
// 	maps := make(map[string]interface{})
// 	data := make(map[string]interface{})

// 	if name != "" {
// 		maps["name"] = name
// 	}

// 	mysql.DB.Model(&models.User{}).Where(maps).
// 		Where("rid > 0").
// 		Offset(util.GetPage(c)).
// 		Limit(util.GetPageSize(c)).
// 		Find(&res)
// 	mysql.DB.Model(&models.User{}).Where(maps).
// 		Where("rid > 0").Count(&count)

// 	data["lists"] = res
// 	data["count"] = count

// 	util.JsonRespond(200, "", data, c)
// }

// // 新增用户
// func PostUser(c *gin.Context) {
// 	if !middleware.PermissionCheckMiddleware(c, "user-add") {
// 		util.JsonRespond(403, "请求资源被拒绝", "", c)
// 		return
// 	}

// 	var user models.User
// 	var data UserResource

// 	if err := c.BindJSON(&data); err != nil {
// 		zap.L().Error("Invalid Add User Data", zap.Error(err))
// 		util.JsonRespond(500, "Invalid Add User Data", "", c)
// 		return
// 	}

// 	// 用户唯一性检查
// 	mysql.DB.Model(&models.User{}).Where("name = ?", data.Name).Find(&user)

// 	if user.ID > 0 {
// 		util.JsonRespond(500, "重复的用户名，请检查！", "", c)
// 		return
// 	}

// 	PasswordHash, err := util.HashPassword(data.Password)
// 	if err != nil {
// 		util.JsonRespond(500, "hash密码错误，请联系管理员！", "", c)
// 		return
// 	}

// 	myuser := models.User{
// 		Name:         data.Name,
// 		Nickname:     data.Nickname,
// 		Mobile:       data.Mobile,
// 		Email:        data.Email,
// 		IsActive:     1,
// 		PasswordHash: PasswordHash,
// 		Rid:          data.Rid}

// 	if err := mysql.DB.Save(&myuser).Error; err != nil {
// 		zap.L().Error("插入数据库有误", zap.Error(err))
// 		util.JsonRespond(500, "插入数据库有误，请联系管理员", "", c)
// 		return
// 	}

// 	util.JsonRespond(200, "添加用户成功", "", c)
// }

// // 删除用户
// func DeleteUser(c *gin.Context) {
// 	if !middleware.PermissionCheckMiddleware(c, "user-del") {
// 		util.JsonRespond(403, "请求资源被拒绝", "", c)
// 		return
// 	}

// 	if err := mysql.DB.Delete(models.User{}, "id = ?", c.Param("id")).Error; err != nil {
// 		zap.L().Error("删除用户失败", zap.Error(err))
// 		util.JsonRespond(500, "删除用户失败", "", c)
// 		return
// 	}
// 	util.JsonRespond(200, "删除成功", "", c)
// }

// // 修改用户
// func PutUser(c *gin.Context) {
// 	if !middleware.PermissionCheckMiddleware(c, "user-edit") {
// 		util.JsonRespond(403, "请求资源被拒绝", "", c)
// 		return
// 	}

// 	var user models.User
// 	var data UserResource

// 	if err := c.BindJSON(&data); err != nil {
// 		zap.L().Error("Invalid Edit User Data", zap.Error(err))
// 		util.JsonRespond(400, "Bad Request : Invalid Edit User Data", "", c)
// 		return
// 	}

// 	mysql.DB.Find(&user, c.Param("id"))
// 	user.Nickname = data.Nickname
// 	user.Mobile = data.Mobile
// 	user.Email = data.Email
// 	user.Rid = data.Rid
// 	user.IsActive = data.IsActive

// 	if len(data.Password) > 0 {
// 		PasswordHash, err := util.HashPassword(data.Password)

// 		if err != nil {
// 			util.JsonRespond(500, "hash密码错误，请联系管理员！", "", c)
// 			return
// 		}

// 		user.PasswordHash = PasswordHash
// 	}

// 	if err := mysql.DB.Save(&user).Error; err != nil {
// 		zap.L().Error("数据写入有问题", zap.Error(err))
// 		util.JsonRespond(500, "数据写入有问题", "", c)
// 		return
// 	}
// 	util.JsonRespond(200, "修改用户成功", "", c)
// }

// // 修改用户
// func PatchUser(c *gin.Context) {
// 	if !middleware.PermissionCheckMiddleware(c, "user-edit") {
// 		util.JsonRespond(403, "请求资源被拒绝", "", c)
// 		return
// 	}

// 	body, err := c.GetRawData()
// 	if err != nil {
// 		zap.L().Error("Invalid Add User Data", zap.Error(err))
// 		util.JsonRespond(500, "Invalid Add User Data", "", c)
// 		return
// 	}

// 	data := make(map[string]string)
// 	json.Unmarshal(body, &data)

// 	uid, _ := c.Get("Uid")
// 	var user models.User
// 	mysql.DB.Model(&models.User{}).Where("id = ?", uid).Find(&user)

// 	switch data["type"] {
// 	case "nickname":
// 		if data["nickname"] == "" {
// 			util.JsonRespond(500, "昵称不能为空！", "", c)
// 			return
// 		}
// 		e := mysql.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]interface{}{"nickname": data["nickname"]}).Error
// 		if e != nil {
// 			util.JsonRespond(500, e.Error(), "", c)
// 			return
// 		}
// 	case "password":
// 		if data["old_password"] == "" {
// 			util.JsonRespond(500, "旧密码不能为空！", "", c)
// 			return
// 		}
// 		if data["new_password"] == "" {
// 			util.JsonRespond(500, "新密码不能为空！", "", c)
// 			return
// 		}
// 		if len(data["new_password"]) < 6 {
// 			util.JsonRespond(500, "请设置至少6位的新密码", "", c)
// 			return
// 		}

// 		if !util.CheckPasswordHash(data["old_password"], user.PasswordHash) {
// 			util.JsonRespond(500, "原密码错误，请重新输入", "", c)
// 			return
// 		}

// 		PasswordHash, e := util.HashPassword(data["new_password"])
// 		if e != nil {
// 			util.JsonRespond(500, "hash密码错误，请联系管理员！", "", c)
// 			return
// 		}

// 		e = mysql.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]interface{}{"password_hash": PasswordHash}).Error
// 		if e != nil {
// 			util.JsonRespond(500, e.Error(), "", c)
// 			return
// 		}

// 	default:
// 		util.JsonRespond(500, "错误的参数", "", c)
// 		return
// 	}

// 	util.JsonRespond(200, "操作成功", "", c)
// }
