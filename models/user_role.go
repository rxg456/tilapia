package models

import "time"

type UserRole struct {
	ID        int    `gorm:"primary_key"`
	Name      string `gorm:"type:varchar(100);not null;default:''"`
	Privilege string `gorm:"type:varchar(2000);not null;default:''"`
	Ctime     int    `gorm:"type:int(11);not null;default:0"`
}

func (m *UserRole) Get(id int) bool {
	return GetByPk(m, id)
}

func (m *UserRole) Count(query QueryParam) (int64, bool) {
	var count int64
	ok := Count(m, &count, query)
	return count, ok
}

func (m *UserRole) Delete() bool {
	return DeleteByPk(m)
}

func (m *UserRole) List(query QueryParam) ([]UserRole, bool) {
	var data []UserRole
	ok := GetMulti(&data, query)
	return data, ok
}

func (m *UserRole) Create() bool {
	m.Ctime = int(time.Now().Unix())
	return Create(m)
}

func (m *UserRole) Update() bool {
	return UpdateByPk(m)
}
