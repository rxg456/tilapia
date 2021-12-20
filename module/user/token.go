package user

import (
	"errors"
	"tilapia/models"
	"time"
)

type Token struct {
	ID     int
	UserId int
	Token  string
	Expire int
}

func (t *Token) DeleteByUserId() error {
	token := &models.UserToken{}
	if ok := token.DeleteByFields(models.QueryParam{
		Where: []models.WhereParam{
			models.WhereParam{
				Field:   "user_id",
				Prepare: t.UserId,
			},
		},
	}); !ok {
		return errors.New("delete user token failed")
	}
	return nil
}

func (t *Token) ValidateToken() bool {
	// if t.UserId == 0 || t.Token == "" {
	// 	return false
	// }
	// TODO
	if t.Token == "" {
		return false
	}

	token := &models.UserToken{}
	if ok := token.GetOne(models.QueryParam{
		Where: []models.WhereParam{
			models.WhereParam{
				Field:   "token",
				Prepare: t.Token,
			},
		},
	}); !ok {
		return false
	}
	if token.ID == 0 {
		return false
	}
	if token.Expire < int(time.Now().Unix()) {
		return false
	}
	return true
}

func (t *Token) CreateOrUpdate() error {
	token := &models.UserToken{}
	if ok := token.GetOne(models.QueryParam{
		Where: []models.WhereParam{
			models.WhereParam{
				Field:   "user_id",
				Prepare: t.UserId,
			},
		},
	}); !ok {
		return errors.New("get user token detail failed")
	}

	token.UserId = t.UserId
	token.Token = t.Token
	token.Expire = t.Expire

	if token.ID == 0 {
		if ok := token.Create(); !ok {
			return errors.New("user token create failed")
		}
	} else {
		if ok := token.Update(); !ok {
			return errors.New("user token update failed")
		}
	}

	return nil
}
