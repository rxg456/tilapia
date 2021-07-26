package models

import (
	"tilapia/dao/mysql"

	"go.uber.org/zap"
)

// 用户
type User struct {
	Model
	Rid          int
	Name         string
	Nickname     string
	PasswordHash string `json:"-"`
	Email        string
	Mobile       string
	IsSupper     int
	IsActive     int
	AccessToken  string
	TokenExpired int64
}

// 菜单权限
type MenuPermissions struct {
	Model
	Pid        int
	Name       string
	Type       int
	Permission string
	Url        string
	Icon       string
	Desc       string
	Children   []*MenuPermissions `json:"children"`
}

func (u *User) ReturnPermissions() []string {
	var res []string
	if u.IsSupper != 1 {
		rows, err := mysql.DB.Table("menu_permissions").
			Select("menu_permissions.permission").
			Joins("left join role_permission_rel on menu_permissions.id = role_permission_rel.pid").
			Where("role_permission_rel.rid = ?", u.Rid).
			Rows()

		if err != nil {
			panic(err)
		}

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				zap.L().Error("ReturnPermissions scan failed", zap.Error(err))
				panic(err)
			}
			res = append(res, name)
		}
	}

	return res
}
