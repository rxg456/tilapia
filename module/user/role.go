package user

import (
	"errors"
	"fmt"
	"tilapia/models"
	"tilapia/pkg/util/gois"
	"tilapia/pkg/util/gostring"
)

type Role struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Privilege []int  `json:"privilege"`
	Ctime     int    `json:"ctime"`
}

func (r *Role) Delete() error {
    role := &models.UserRole{
        ID: r.ID,
    }
    if ok := role.Delete(); !ok {
        return errors.New("delete user role failed")
    }
    return nil
}

func (r *Role) CreateOrUpdate() error {
	role := &models.UserRole{
		ID:        r.ID,
		Name:      r.Name,
		Privilege: gostring.JoinIntSlice2String(r.Privilege, ","),
	}
	if role.ID == 0 {
		if ok := role.Create(); !ok {
			return errors.New("create user role data failed")
		}
	} else {
		if ok := role.Update(); !ok {
			return errors.New("update user role failed")
		}
	}
	return nil
}

func (r *Role) Total(keyword string) (int64, error) {
	role := &models.UserRole{}
	total, ok := role.Count(models.QueryParam{
		Where: r.parseWhereConds(keyword),
	})
	if !ok {
		return 0, errors.New("get user role count failed")
	}
	return total, nil
}

func (r *Role) List(keyword string, offset, limit int) ([]Role, error) {
	role := &models.UserRole{}
	list, ok := role.List(models.QueryParam{
		Fields: "id, name, ctime",
		Offset: offset,
		Limit:  limit,
		Order:  "id ASC",
		Where:  r.parseWhereConds(keyword),
	})
	if !ok {
		return nil, errors.New("get user role list failed")
	}

	var roleList []Role
	for _, l := range list {
		roleList = append(roleList, Role{
			ID:    l.ID,
			Name:  l.Name,
			Ctime: l.Ctime,
		})
	}
	return roleList, nil
}

func (r *Role) parseWhereConds(keyword string) []models.WhereParam {
	var where []models.WhereParam
	if keyword != "" {
		if gois.IsInteger(keyword) {
			where = append(where, models.WhereParam{
				Field:   "id",
				Prepare: keyword,
			})
		} else {
			where = append(where, models.WhereParam{
				Field:   "name",
				Tag:     "LIKE",
				Prepare: fmt.Sprintf("%%%s%%", keyword),
			})
		}
	}
	return where
}

func RoleGetMapByIds(ids []int) (map[int]Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	role := &models.UserRole{}
	list, ok := role.List(models.QueryParam{
		Where: []models.WhereParam{
			models.WhereParam{
				Field:   "id",
				Tag:     "IN",
				Prepare: ids,
			},
		},
	})
	if !ok {
		return nil, errors.New("get user role maps failed")
	}
	roleMap := make(map[int]Role)
	for _, l := range list {
		roleMap[l.ID] = Role{
			ID:   l.ID,
			Name: l.Name,
		}
	}
	return roleMap, nil
}

func (r *Role) Detail() error {
	role := &models.UserRole{}
	if ok := role.Get(r.ID); !ok {
		return errors.New("get user role detail failed")
	}
	if role.ID == 0 {
		return errors.New("user role not exists")
	}

	r.ID = role.ID
	r.Name = role.Name
	r.Privilege = gostring.StrSplit2IntSlice(role.Privilege, ",")
	r.Ctime = role.Ctime

	return nil
}
