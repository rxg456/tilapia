package models

import (
	"time"
)

type UserToken struct {
	ID     int    `gorm:"primary_key"`
	UserId int    `gorm:"type:int(11);not null;default:0"`
	Token  string `gorm:"type:varchar(100);not null;default:''"`
	Expire int    `gorm:"type:int(11);not null;default:0"`
	Ctime  int    `gorm:"type:int(11);not null;default:0"`
}

func (m *UserToken) GetOne(query QueryParam) bool {
	return GetOne(m, query)
}

func (m *UserToken) Create() bool {
	m.Ctime = int(time.Now().Unix())
	return Create(m)
}

func (m *UserToken) Update() bool {
	return UpdateByPk(m)
}

func (m *UserToken) UpdateByFields(data map[string]interface{}, query QueryParam) bool {
	return Update(m, data, query)
}

func (m *UserToken) DeleteByFields(query QueryParam) bool {
	return Delete(m, query)
}
