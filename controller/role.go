package controller

import (
	"fmt"
	"strconv"
	"tilapia/dao/mysql"
	"tilapia/dao/redis"
	"tilapia/middleware"
	"tilapia/models"
	"tilapia/pkg/util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleResource struct {
	Name string `form:"Name"`
	Desc string `form:"Desc"`
}

type RolePermResource struct {
	Codes []int `json:Codes`
}

// 角色列表
func GetRole(c *gin.Context) {
	var roles []models.Role
	var count int64
	maps := make(map[string]interface{})
	data := make(map[string]interface{})

	mysql.DB.Model(&models.Role{}).
		Where(maps).
		Offset(util.GetPage(c)).
		Limit(util.GetPageSize(c)).
		Find(&roles)
	mysql.DB.Model(&models.Role{}).
		Where(maps).
		Count(&count)

	data["lists"] = roles
	data["count"] = count
	util.JsonRespond(200, "", data, c)
}

// 新增用户角色
func PostRole(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "role-add") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var data RoleResource
	var role models.Role
	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid Add Role Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Add Role Data", "", c)
		return
	}

	// 角色名唯一性检查
	mysql.DB.Model(&models.Role{}).
		Where("name = ?", data.Name).
		Find(&role)

	if role.ID > 0 {
		util.JsonRespond(500, "重复的角色名，请检查！", "", c)
		return
	}

	role = models.Role{
		Name: data.Name,
		Desc: data.Desc,
	}

	if err := mysql.DB.Save(&role).Error; err != nil {
		zap.L().Error("Role Save faild", zap.Error(err))
		util.JsonRespond(500, "内部错误", "", c)
		return
	}

	util.JsonRespond(200, "添加角色成功", "", c)
}

// 删除角色
func DeleteRole(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "role-del") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}
	if err := mysql.DB.Delete(models.Role{}, "id = ?", c.Param("id")).Error; err != nil {
		zap.L().Error("Delete Role db faild", zap.Error(err))
		util.JsonRespond(500, "内部错误", "", c)
		return
	}

	util.JsonRespond(200, "删除成功", "", c)
}

// 修改角色
func PutRole(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "role-edit") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var data RoleResource
	var role models.Role

	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid Edit Role Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Edit Role Data", "", c)
		return
	}

	// 角色名唯一性检查
	mysql.DB.Model(&models.Role{}).
		Where("name = ?", data.Name).
		Where("id != ?", c.Param("id")).
		Find(&role)

	if role.ID > 0 {
		util.JsonRespond(500, "重复的角色名，请检查！", "", c)
		return
	}

	mysql.DB.Find(&role, c.Param("id"))

	role.Name = data.Name
	role.Desc = data.Desc

	if err := mysql.DB.Save(&role).Error; err != nil {
		zap.L().Error("role save faild", zap.Error(err))
		util.JsonRespond(500, "role save faild", "", c)
		return
	}

	util.JsonRespond(200, "修改角色成功", "", c)
}

// 角色权限详情
func GetRolePerms(c *gin.Context) {
	var data map[string]interface{}
	var prole []models.RolePermissionRel
	data = make(map[string]interface{})

	mysql.DB.Model(&models.RolePermissionRel{}).
		Select("pid").
		Where("rid = ?", c.Param("id")).
		Find(&prole)

	data["lists"] = prole

	util.JsonRespond(200, "", data, c)
}

// 添加/修改角色功能权限
func PostRolePerms(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "role-perm-add") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var data RolePermResource
	var old_prole []models.RolePermissionRel
	var rpr models.RolePermissionRel
	var mps []models.MenuPermissions

	rds := make(map[int]interface{})
	prole := make(map[int]interface{})
	new_prole := make(map[int]interface{})
	rid, _ := strconv.Atoi(c.Param("id"))

	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid MenuPermissions Data", zap.Error(err))
		util.JsonRespond(500, "Invalid MenuPermissions Data", "", c)
		return
	}

	mysql.DB.Model(&models.RolePermissionRel{}).
		Select("pid").
		Where("rid = ?", rid).
		Find(&old_prole)

	// 可以把所有的 type=1 的菜单选项id 放到 rds队列里
	mysql.DB.Model(&models.MenuPermissions{}).
		Select("id").
		Where("type = ?", 1).
		Find(&mps)

	for _, v := range mps {
		rds[v.ID] = v.ID
	}

	for _, v := range data.Codes {
		//m, _ := strconv.Atoi(v)
		if _, ok := rds[v]; ok {
			continue
		}

		new_prole[v] = v
	}

	// 删除
	for _, k := range old_prole {
		prole[k.Pid] = k.Pid
		if _, ok := new_prole[k.Pid]; !ok {
			// 执行删除操作
			if err := mysql.DB.Delete(models.RolePermissionRel{}, "pid = ?", k.Pid).Error; err != nil {
				zap.L().Error("mysql Delete roleperms faild", zap.Error(err))
				util.JsonRespond(500, "Delete roleperms faild", "", c)
				return
			}
		}
	}

	// 新增
	for k, _ := range new_prole {
		if _, ok := prole[k]; !ok {
			//执行新增操作，换成批量插入更好
			rpr = models.RolePermissionRel{
				Pid: k,
				Rid: rid}

			if err := mysql.DB.Save(&rpr).Error; err != nil {
				zap.L().Error("mysql add roleperms faild", zap.Error(err))
				util.JsonRespond(500, "mysql add roleperms faild", "", c)
				return
			}
		}
	}

	//更新redis里面的角色的权限集合
	key := redis.RoleRermsSetKey
	str := fmt.Sprintf("%v", rid)
	key = key + str

	// 先删除key
	redis.DelKey(key)
	// 再添加
	models.SetRolePermToSet(key, rid)

	util.JsonRespond(200, "", "", c)
}
