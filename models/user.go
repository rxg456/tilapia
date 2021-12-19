package models

import "time"

// 用户
type User struct {
	ID            int    `gorm:"primary_key"`
	RoleId        int    `gorm:"type:int(11);not null;default:0"`
	Username      string `gorm:"type:varchar(20);not null;default:''"`
	Password      string `gorm:"type:char(32);not null;default:''"`
	Truename      string `gorm:"type:varchar(20);not null;default:''"`
	Mobile        string `gorm:"type:varchar(20);not null;default:''"`
	Email         string `gorm:"type:varchar(500);not null;default:''"`
	Status        int    `gorm:"type:int(11);not null;default:0"`
	LastLoginTime int    `gorm:"type:int(11);not null;default:0"`
	LastLoginIp   string `gorm:"type:varchar(50);not null;default:''"`
	Ctime         int    `gorm:"type:int(11);not null;default:0"`
}

func (m *User) Delete() bool {
    return DeleteByPk(m)
}

func (m *User) List(query QueryParam) ([]User, bool) {
	var data []User
	ok := GetMulti(&data, query)
	return data, ok
}

func (m *User) GetOne(query QueryParam) bool {
	return GetOne(m, query)
}

func (m *User) UpdateByFields(data map[string]interface{}, query QueryParam) bool {
	return Update(m, data, query)
}

func (m *User) Count(query QueryParam) (int64, bool) {
	var count int64
	ok := Count(m, &count, query)
	return count, ok
}

func (m *User) Create() bool {
	m.Ctime = int(time.Now().Unix())
	return Create(m)
}

// // 菜单权限
// type MenuPerms struct {
// 	ID         int `gorm:"primary_key" json:"id"`
// 	Pid        int
// 	Name       string
// 	Type       int
// 	Permission string
// 	Url        string
// 	Icon       string
// 	Desc       string
// 	Children   []MenuPermissions `json:"children"`
// }

// // 递归菜单权限
// type MenuPermissions struct {
// 	ID         int `gorm:"primary_key" json:"id"`
// 	Pid        int
// 	Name       string
// 	Type       int
// 	Permission string
// 	Url        string
// 	Icon       string
// 	Desc       string
// }

// func (u *User) ReturnPermissions() []string {
// 	var res []string
// 	if u.IsSupper != 1 {
// 		rows, err := mysql.DB.Table("menu_permissions").
// 			Select("menu_permissions.permission").
// 			Joins("left join role_permission_rel on menu_permissions.id = role_permission_rel.pid").
// 			Where("role_permission_rel.rid = ?", u.Rid).
// 			Rows()

// 		if err != nil {
// 			panic(err)
// 		}

// 		for rows.Next() {
// 			var name string
// 			if err := rows.Scan(&name); err != nil {
// 				zap.L().Error("ReturnPermissions scan failed", zap.Error(err))
// 				panic(err)
// 			}
// 			res = append(res, name)
// 		}
// 	}

// 	return res
// }

// func SetRolePermToSet(key string, rid int) {
// 	var mps []MenuPermissions

// 	mysql.DB.Table("menu_permissions").
// 		Select("menu_permissions.permission").
// 		Joins("left join role_permission_rels on menu_permissions.id = role_permission_rels.pid").
// 		Where("role_permission_rels.rid = ?", rid).
// 		Find(&mps)

// 	for _, v := range mps {
// 		err := redis.SetValBySetKey(key, v.Permission)
// 		if err != nil {
// 			zap.L().Error("SetRolePermToSet faild", zap.Error(err))
// 		}
// 	}
// }
