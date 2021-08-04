package controller

import (
	"sort"
	"tilapia/dao/mysql"
	"tilapia/dao/redis"
	"tilapia/middleware"
	"tilapia/models"
	"tilapia/pkg/util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PermResource struct {
	Name       string `form:"Name"`
	Permission string `form:"Permission"`
	Type       int    `form:"Type"`
	Pid        int    `form:"Pid"`
	Desc       string `form:"Desc"`
}

// 权限列表
func GetPerms(c *gin.Context) {
	var mps []models.MenuPermissions
	var count int64
	maps := make(map[string]interface{})
	data := make(map[string]interface{})

	mysql.DB.Model(&models.MenuPermissions{}).Where(maps).Where("type = ?", 2).
		Offset(util.GetPage(c)).
		Limit(util.GetPageSize(c)).
		Find(&mps)

	mysql.DB.Model(&models.MenuPermissions{}).Where(maps).Where("type = ?", 2).Count(&count)
	data["lists"] = mps
	data["count"] = count
	util.JsonRespond(200, "", data, c)
}

// 权限添加
func PostPerms(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "perm-add") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var data PermResource
	var mps models.MenuPermissions

	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid Add Perm Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Add Perm Data", "", c)
		return
	}

	// 权限项唯一性检查
	mysql.DB.Model(&models.MenuPermissions{}).
		Where("permission = ?", data.Permission).
		Where("type = ?", 2).
		Find(&mps)

	if mps.ID > 0 {
		zap.L().Error("重复性的标识符")
		util.JsonRespond(500, "重复性的标识符，请检查！", "", c)
		return
	}

	perm := models.MenuPermissions{
		Name:       data.Name,
		Permission: data.Permission,
		Desc:       data.Desc,
		Pid:        data.Pid,
		Type:       data.Type,
	}
	if err := mysql.DB.Save(&perm).Error; err != nil {
		zap.L().Error("perm save faild", zap.Error(err))
		util.JsonRespond(500, "内部错误", "", c)
		return
	}

	// 修改perm信息，删除redis的key值
	redis.DelRedisAllPermKey()

	util.JsonRespond(200, "添加菜单权限按钮成功", "", c)
}

// 权限修改
func PutPerms(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "perm-edit") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	var data PermResource
	var perm models.MenuPermissions
	var mps models.MenuPermissions

	if err := c.BindJSON(&data); err != nil {
		zap.L().Error("Invalid Edit Perm Data", zap.Error(err))
		util.JsonRespond(500, "Invalid Edit Perm Data", "", c)
		return
	}

	// 权限项唯一性检查
	mysql.DB.Model(&models.MenuPermissions{}).
		Where("permission = ?", data.Permission).
		Where("id != ?", c.Param("id")).
		Find(&mps)

	if mps.ID > 0 {
		zap.L().Error("重复性的标识符")
		util.JsonRespond(500, "重复性的标识符，请检查！", "", c)
		return
	}

	mysql.DB.Find(&perm, c.Param("id"))

	perm.Name = data.Name
	perm.Desc = data.Desc
	perm.Pid = data.Pid
	perm.Permission = data.Permission

	err := mysql.DB.Save(&perm).Error
	if err != nil {
		util.JsonRespond(500, "内部错误", "", c)
		return
	}

	// 修改perm 信息， 最简单的做法， 删除redis 对应的key
	redis.DelRedisAllPermKey()

	util.JsonRespond(200, "修改权限按钮成功", "", c)

}

// 删除权限按钮
func DeletePerms(c *gin.Context) {
	if !middleware.PermissionCheckMiddleware(c, "perm-del") {
		util.JsonRespond(403, "请求资源被拒绝", "", c)
		return
	}

	if err := mysql.DB.Delete(models.MenuPermissions{}, "id = ?", c.Param("id")).Error; err != nil {
		zap.L().Error("delete perms faild", zap.Error(err))
		util.JsonRespond(500, "内部错误", "", c)
		return
	}
	// 删除缓存
	redis.DelRedisAllPermKey()

	util.JsonRespond(200, "删除权限按钮成功", "", c)

}

// 获取所有的权限项
func GetAllPerms(c *gin.Context) {
	var perms []models.MenuPermissions
	var res util.SortMenuPerms

	data := make(map[string]interface{})
	tmp := make(map[int]*models.MenuPerms)

	// 所有的mod page perm组合数据 放到redis里面
	key := redis.AllPermsKey
	v, _ := redis.Rdb.Get(key).Result()

	if v != "" {
		data["lists"] = util.JsonUnmarshalFromString(v, &res)
		util.JsonRespond(200, "", data, c)
		return
	}

	mysql.DB.Model(&models.MenuPermissions{}).Find(&perms)

	for _, p := range perms {
		if p.Pid == 0 {
			tmp[p.ID] = &models.MenuPerms{
				ID:         p.ID,
				Pid:        p.Pid,
				Name:       p.Name,
				Type:       p.Type,
				Permission: p.Permission,
				Url:        p.Url,
				Icon:       p.Icon,
				Desc:       p.Desc,
			}
		} else {
			if x, ok := tmp[p.Pid]; ok {
				x.Children = append(x.Children, p)
			} else {
				tmp[p.Pid] = &models.MenuPerms{
					Children: []models.MenuPermissions{p},
				}
			}
		}
	}
	for _, v := range tmp {
		if v.Pid == 0 {
			res = append(res, v)
		}
	}
	sort.Stable(res)

	redis.Rdb.Set(key, util.JSONMarshalToString(res), 0)

	data["lists"] = res

	util.JsonRespond(200, "", data, c)
}
