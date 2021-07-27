package controller

import (
	"encoding/json"
	"fmt"
	"tilapia/dao/mysql"
	"tilapia/dao/redis"
	"tilapia/middleware"
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
}

type UserResource struct {
	Name     string `form:"Name"`
	Nickname string `form:"Nickname"`
	Mobile   string `form: Mobile`
	Email    string `form: Email`
	Rid      int    `form: Rid`
	Password string `form:"password"`
	IsActive int    `form:IsActive`
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

func Logout(c *gin.Context) {
	var user models.User

	Uid, _ := c.Get("Uid")
	mysql.DB.Find(&user, Uid)
	user.AccessToken = ""
	err := mysql.DB.Save(&user).Error
	if err != nil {
		zap.L().Error("logout db save faild", zap.Error(err))
		util.JsonRespond(500, err.Error(), "", c)
		return
	}
	util.JsonRespond(200, "退出成功", "", c)
}

func GetUserMenu(c *gin.Context) {
	var mps []*models.MenuPermissions
	var res []models.MenuPermissions

	tmp := make(map[int]*models.MenuPermissions)
	data := make(map[string]interface{})
	rid := c.Param("id")

	// 用户菜单列表
	key := redis.RoleMenuListKey
	str := fmt.Sprintf("%v", rid)
	key = key + str
	v, err := redis.Rdb.Get(key).Result()
	if err != nil {
		zap.L().Error("rolemenulist get faild", zap.Error(err))
	}
	if v != "" {
		data["lists"] = util.JsonUnmarshalFromString(v, &res)
		util.JsonRespond(200, "", data, c)
		return
	}
	// 超级用户返回所有菜单
	if rid == "0" {
		mysql.DB.Model(&models.MenuPermissions{}).Where("type = 1").Find(&mps)
		for _, p := range mps {
			if x, ok := tmp[p.ID]; ok {
				p.Children = x.Children
			}
			tmp[p.ID] = p
			if p.Pid != 0 {
				if x, ok := tmp[p.Pid]; ok {
					x.Children = append(x.Children, p)
				} else {
					tmp[p.Pid] = &models.MenuPermissions{
						Children: []*models.MenuPermissions{p},
					}
				}
			}
		}
	} else {
		// 普通用户 根据 rid 返回菜单项
		pids := []int{}
		mysql.DB.Table("menu_permissions").
			Select("menu_permissions.pid").
			Joins("left join role_permission_rel on menu_permissions.id = role_permission_rel.pid").
			Where("role_permission_rel.rid = ?", rid).
			Pluck("DISTINCT menu_permissions.pid", &pids)

		mysql.DB.Model(&models.MenuPermissions{}).
			Where("type = ?", 1).
			Find(&mps)

		for _, v := range mps {
			for _, p := range pids {
				if _, ok := tmp[v.ID]; !ok {
					tmp[v.ID] = v
				}

				if p == v.ID {
					if x, ok := tmp[v.Pid]; ok {
						x.Children = append(x.Children, v)
					}
				}
			}
		}
	}

	for _, v := range tmp {
		if v.Pid == 0 {
			res = append(res, *v)
		}
	}

	redis.Rdb.Set(key, util.JSONMarshalToString(res), 0)

	data["lists"] = res
	util.JsonRespond(200, "", data, c)
}

// 用户列表
func GetUsers(c *gin.Context) {
	var res []models.User
	var count int64

	name := c.Query("name")
	maps := make(map[string]interface{})
	data := make(map[string]interface{})

	if name != "" {
		maps["name"] = name
	}

	mysql.DB.Model(&models.User{}).Where(maps).
		Where("rid > 0").
		Offset(util.GetPage(c)).
		Limit(util.GetPageSize(c)).
		Find(&res)
	mysql.DB.Model(&models.User{}).Where(maps).
		Where("rid > 0").Count(&count)

	data["lists"] = res
	data["count"] = count

	util.JsonRespond(200, "", data, c)
}

// 新增用户
func PostUser(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "user-add") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var user models.User
	var data UserResource

	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid Add User Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Add User Data", "", c)
		return
	}

	// 用户唯一性检查
	mysql.DB.Model(&models.User{}).Where("name = ?", data.Name).Find(&user)

	if user.ID > 0 {
		util.JsonRespond(500, "重复的用户名，请检查！", "", c)
		return
	}

	PasswordHash, err := util.HashPassword(data.Password)
	if err != nil {
		util.JsonRespond(500, "hash密码错误，请联系管理员！", "", c)
		return
	}

	myuser := models.User{
		Name:         data.Name,
		Nickname:     data.Nickname,
		Mobile:       data.Mobile,
		Email:        data.Email,
		IsActive:     1,
		PasswordHash: PasswordHash,
		Rid:          data.Rid}

	if err := mysql.DB.Save(&myuser).Error; err != nil {
		zap.L().Error("插入数据库有误", zap.Error(err))
		util.JsonRespond(500, "插入数据库有误，请联系管理员", "", c)
		return
	}

	util.JsonRespond(200, "添加用户成功", "", c)
}

// 删除用户
func DeleteUser(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "user-del") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	if err := mysql.DB.Delete(models.User{}, "id = ?", c.Param("id")).Error; err != nil {
		zap.L().Error("删除用户失败", zap.Error(err))
		util.JsonRespond(500, "删除用户失败", "", c)
		return
	}
	util.JsonRespond(200, "删除成功", "", c)
}

// 修改用户
func PutUser(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "user-edit") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var user models.User
	var data UserResource

	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid Edit User Data", zap.Error(err))
		util.JsonRespond(400, "Bad Request : Invalid Edit User Data", "", c)
		return
	}

	mysql.DB.Find(&user, c.Param("id"))
	user.Nickname = data.Nickname
	user.Mobile = data.Mobile
	user.Email = data.Email
	user.Rid = data.Rid
	user.IsActive = data.IsActive

	if len(data.Password) > 0 {
		PasswordHash, err := util.HashPassword(data.Password)

		if err != nil {
			util.JsonRespond(500, "hash密码错误，请联系管理员！", "", c)
			return
		}

		user.PasswordHash = PasswordHash
	}

	if err := mysql.DB.Save(&user).Error; err != nil {
		zap.L().Error("数据写入有问题", zap.Error(err))
		util.JsonRespond(500, "数据写入有问题", "", c)
		return
	}
	util.JsonRespond(200, "修改用户成功", "", c)
}

// 修改用户
func PatchUser(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "user-edit") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		zap.L().Error("Invalid Add User Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Add User Data", "", c)
		return
	}

	data := make(map[string]string)
	json.Unmarshal(body, &data)

	uid, _ := c.Get("Uid")
	var user models.User
	mysql.DB.Model(&models.User{}).Where("id = ?", uid).Find(&user)

	switch data["type"] {
	case "nickname":
		if data["nickname"] == "" {
			util.JsonRespond(500, "昵称不能为空！", "", c)
			return
		}
		e := mysql.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]interface{}{"nickname": data["nickname"]}).Error
		if e != nil {
			util.JsonRespond(500, e.Error(), "", c)
			return
		}
	case "password":
		if data["old_password"] == "" {
			util.JsonRespond(500, "旧密码不能为空！", "", c)
			return
		}
		if data["new_password"] == "" {
			util.JsonRespond(500, "新密码不能为空！", "", c)
			return
		}
		if len(data["new_password"]) < 6 {
			util.JsonRespond(500, "请设置至少6位的新密码", "", c)
			return
		}

		if !util.CheckPasswordHash(data["old_password"], user.PasswordHash) {
			util.JsonRespond(500, "原密码错误，请重新输入", "", c)
			return
		}

		PasswordHash, e := util.HashPassword(data["new_password"])
		if e != nil {
			util.JsonRespond(500, "hash密码错误，请联系管理员！", "", c)
			return
		}

		e = mysql.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]interface{}{"password_hash": PasswordHash}).Error
		if e != nil {
			util.JsonRespond(500, e.Error(), "", c)
			return
		}

	default:
		util.JsonRespond(500, "错误的参数", "", c)
		return
	}

	util.JsonRespond(200, "操作成功", "", c)
}
