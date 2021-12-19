package models

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"tilapia/dao/mysql"
)

type WhereParam struct {
	Field   string
	Tag     string
	Prepare interface{}
}

type QueryParam struct {
	Fields string // 字段
	Offset int
	Limit  int
	Order  string
	Where  []WhereParam
}

type Model struct {
	ID int `gorm:"primary_key" json:"id"`
}

func DeleteByPk(model interface{}) bool {
	db := mysql.DB.Model(model)
	db.Delete(model)
	if err := db.Error; err != nil {
		zap.L().Error("mysql query error:", zap.Error(err))
		return false
	}
	return true
}

func GetMulti(model interface{}, query QueryParam) bool {
	db := mysql.DB.Offset(query.Offset)
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Fields != "" {
		db = db.Select(query.Fields)
	}
	if query.Order != "" {
		db = db.Order(query.Order)
	}
	db = parseWhereParam(db, query.Where)
	db.Find(model)
	if err := db.Error; err != nil {
		zap.L().Error("mysql query error:", zap.Error(err))
		return false
	}
	return true
}

func Count(model interface{}, count *int64, query QueryParam) bool {
	db := mysql.DB.Model(model)
	db = parseWhereParam(db, query.Where)
	db = db.Count(count)
	if err := db.Error; err != nil {
		zap.L().Error("mysql execute error:", zap.Error(err))
		return false
	}
	return true
}
func Create(model interface{}) bool {
	db := mysql.DB.Create(model)
	if err := db.Error; err != nil {
		zap.L().Error("mysql execute error:", zap.Error(err))
		return false
	}
	return true
}

func Delete(model interface{}, query QueryParam) bool {
	if len(query.Where) == 0 {
		zap.L().Warn("mysql query error: delete failed, where conditions cannot be empty")
		return false
	}
	db := mysql.DB.Model(model)
	db = parseWhereParam(db, query.Where)
	db = db.Delete(model)
	if err := db.Error; err != nil {
		zap.L().Warn("mysql query error: delete failed, where conditions cannot be empty", zap.Error(err))
		return false
	}
	return true
}

func GetByPk(model interface{}, id interface{}) bool {
	db := mysql.DB.Model(model)
	db.First(model, id)
	if err := db.Error; err != nil {
		zap.L().Error("mysql query error:", zap.Error(err))
		return false
	}
	return true
}

func Update(model interface{}, data interface{}, query QueryParam) bool {
	db := mysql.DB.Model(model)
	db = parseWhereParam(db, query.Where)
	db = db.Updates(data)
	if err := db.Error; err != nil {
		zap.L().Error("mysql query error:", zap.Error(err))
		return false
	}
	return true
}

func UpdateByPk(model interface{}) bool {
	db := mysql.DB.Model(model)
	db = db.Updates(model)
	if err := db.Error; err != nil {
		zap.L().Error("mysql query error:", zap.Error(err))
		return false
	}
	return true
}

func GetOne(model interface{}, query QueryParam) bool {
	db := mysql.DB.Model(model)
	if query.Fields != "" {
		db = db.Select(query.Fields)
	}
	db = parseWhereParam(db, query.Where)
	db = db.First(model)
	// if err := db.Error; err != nil {
	// 	zap.L().Error("用户不存在:", zap.Error(err))
	// 	return false
	// }
	return true
}

func parseWhereParam(db *gorm.DB, where []WhereParam) *gorm.DB {
	if len(where) == 0 {
		return db
	}
	var (
		plain   []string
		prepare []interface{}
	)
	for _, w := range where {
		tag := w.Tag
		if tag == "" {
			tag = "="
		}
		var plainFmt string
		switch tag {
		case "IN":
			plainFmt = fmt.Sprintf("%s IN (?)", w.Field)
		default:
			plainFmt = fmt.Sprintf("%s %s ?", w.Field, tag)
		}
		plain = append(plain, plainFmt)
		prepare = append(prepare, w.Prepare)
	}
	return db.Where(strings.Join(plain, " AND "), prepare...)
}
