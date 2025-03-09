package rule

import (
	"fmt"

	"gorm.io/gorm"
)

// NotExists 验证一个值在某个表中的字段中不存在，支持同时判断多个字段
// NotExists verify a value does not exist in a table field, support judging multiple fields at the same time
// 用法：notExists:表名称,字段名称,字段名称,字段名称
// Usage: notExists:table_name,field_name,field_name,field_name
// 例子：notExists:users,phone,email
// Example: notExists:users,phone,email
type NotExists struct {
	db *gorm.DB
}

func NewNotExists(db *gorm.DB) *NotExists {
	return &NotExists{db: db}
}

func (r *NotExists) Passes(val any, options ...any) bool {
	if len(options) < 2 {
		return false
	}

	tableName := options[0].(string)
	fieldNames := options[1:]

	query := r.db.Table(tableName).Where(fmt.Sprintf("%s = ?", fieldNames[0]), val)
	for _, fieldName := range fieldNames[1:] {
		query = query.Or(fmt.Sprintf("%s = ?", fieldName), val)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false
	}

	return count == 0
}
