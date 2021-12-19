package user

import (
	"errors"
	"fmt"

	"tilapia/models"
	"tilapia/pkg/util"
	"tilapia/pkg/util/gois"

	"go.uber.org/zap"
)

type User struct {
	ID            int    `json:"id"`
	RoleId        int    `json:"role_id"`
	RoleName      string `json:"role_name"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	Truename      string `json:"truename"`
	Mobile        string `json:"mobile"`
	Status        int    `json:"status"`
	LastLoginTime int    `json:"last_login_time"`
	LastLoginIp   string `json:"last_login_ip"`
	Ctime         int    `json:"ctime"`
}

func (u *User) Delete() error {
    user := &models.User{
        ID: u.ID,
    }
    if ok := user.Delete(); !ok {
        return errors.New("user delete failed")
    }
    return nil
}

func (u *User) Total(keyword string) (int64, error) {
	user := &models.User{}
	total, ok := user.Count(models.QueryParam{
		Where: u.parseWhereConds(keyword),
	})
	if !ok {
		return 0, errors.New("get user list count failed")
	}
	return total, nil
}

func (u *User) List(keyword string, offset, limit int) ([]User, error) {
	user := &models.User{}
	list, ok := user.List(models.QueryParam{
		Fields: "id, role_id, username, email, status, last_login_time, last_login_ip",
		Offset: offset,
		Limit:  limit,
		Order:  "id DESC",
		Where:  u.parseWhereConds(keyword),
	})
	if !ok {
		return nil, errors.New("get user list failed")
	}
	var roleIdList []int
	for _, l := range list {
		roleIdList = append(roleIdList, l.RoleId)
	}
	roleMap, err := RoleGetMapByIds(roleIdList)
	if err != nil {
		return nil, errors.New("get user map list failed")
	}

	var userList []User
	for _, l := range list {
		user := User{
			ID:            l.ID,
			RoleId:        l.RoleId,
			Username:      l.Username,
			Email:         l.Email,
			Status:        l.Status,
			LastLoginTime: l.LastLoginTime,
			LastLoginIp:   l.LastLoginIp,
		}
		if r, exists := roleMap[user.RoleId]; exists {
			user.RoleName = r.Name
		}
		userList = append(userList, user)
	}
	return userList, nil
}

func (u *User) parseWhereConds(keyword string) []models.WhereParam {
	var where []models.WhereParam
	if keyword != "" {
		if gois.IsInteger(keyword) {
			where = append(where, models.WhereParam{
				Field:   "id",
				Prepare: keyword,
			})
		} else {
			if gois.IsEmail(keyword) {
				where = append(where, models.WhereParam{
					Field:   "email",
					Prepare: keyword,
				})
			} else {
				where = append(where, models.WhereParam{
					Field:   "username",
					Tag:     "LIKE",
					Prepare: fmt.Sprintf("%%%s%%", keyword),
				})
			}
		}
	}
	return where
}

func (u *User) CreateOrUpdate() error {
	var password string
	var err error
	if u.Password != "" {
		password, err = util.HashPassword(u.Password)
		if err != nil {
			zap.L().Error("加密失败", zap.Error(err))
		}
	}
	user := &models.User{
		ID:       u.ID,
		RoleId:   u.RoleId,
		Username: u.Username,
		Email:    u.Email,
		Truename: u.Truename,
		Mobile:   u.Mobile,
		Status:   u.Status,
	}
	if u.ID > 0 {
		updateData := map[string]interface{}{
			"role_id":  u.RoleId,
			"username": u.Username,
			"email":    u.Email,
			"truename": u.Truename,
			"mobile":   u.Mobile,
			"status":   u.Status,
		}
		if password != "" {
			updateData["password"] = password
		}
		ok := user.UpdateByFields(updateData, models.QueryParam{
			Where: []models.WhereParam{
				models.WhereParam{
					Field:   "id",
					Prepare: u.ID,
				},
			},
		})
		if !ok {
			return errors.New("user update failed")
		}
	} else {
		user.Password = password
		ok := user.Create()
		if !ok {
			return errors.New("user update failed")
		}
	}
	return nil
}

func (u *User) UserCheckExists() (bool, error) {
	var where []models.WhereParam
	if u.Username != "" {
		where = append(where, models.WhereParam{
			Field:   "username",
			Prepare: u.Username,
		})
	}
	if u.ID != 0 {
		where = append(where, models.WhereParam{
			Field:   "id",
			Tag:     "!=",
			Prepare: u.ID,
		})
	}
	if u.Email != "" {
		where = append(where, models.WhereParam{
			Field:   "email",
			Prepare: u.Email,
		})
	}
	user := &models.User{}
	count, ok := user.Count(models.QueryParam{
		Where: where,
	})
	if !ok {
		return false, errors.New("check user failed")
	}

	return count > 0, nil
}

func (u *User) UpdatePassword() error {
	user := &models.User{}

	PasswordHash, err := util.HashPassword(u.Password)
	if err != nil {
		return errors.New("hash密码错误，请联系管理员！")
	}
	updateData := map[string]interface{}{
		"password": PasswordHash,
	}
	ok := user.UpdateByFields(updateData, models.QueryParam{
		Where: []models.WhereParam{
			models.WhereParam{
				Field:   "id",
				Prepare: u.ID,
			},
		},
	})
	if !ok {
		return errors.New("user password update failed")
	}
	return nil
}

func (u *User) UserSettingUpdate() error {
	user := &models.User{}
	updateData := map[string]interface{}{
		"truename": u.Truename,
		"mobile":   u.Mobile,
	}
	ok := user.UpdateByFields(updateData, models.QueryParam{
		Where: []models.WhereParam{
			models.WhereParam{
				Field:   "id",
				Prepare: u.ID,
			},
		},
	})
	if !ok {
		return errors.New("user update failed")
	}
	return nil
}

func (u *User) Detail() error {
	var where []models.WhereParam
	user := &models.User{}

	if u.ID != 0 {
		where = append(where, models.WhereParam{
			Field:   "id",
			Prepare: u.ID,
		})
	}
	if u.Username != "" {
		where = append(where, models.WhereParam{
			Field:   "username",
			Prepare: u.Username,
		})
	}
	if u.Email != "" {
		where = append(where, models.WhereParam{
			Field:   "email",
			Prepare: u.Email,
		})
	}
	if ok := user.GetOne(models.QueryParam{
		Where: where,
	}); !ok {
		return errors.New("get user detail failed")
	}
	if user.ID == 0 {
		return errors.New("user not exists")
	}

	u.ID = user.ID
	u.RoleId = user.RoleId
	u.Username = user.Username
	u.Password = user.Password
	u.Email = user.Email
	u.Truename = user.Truename
	u.Mobile = user.Mobile
	u.Status = user.Status
	u.LastLoginTime = user.LastLoginTime
	u.LastLoginIp = user.LastLoginIp
	u.Ctime = user.Ctime

	return nil
}
