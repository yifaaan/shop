package handler

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// 分页查询
func paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func generateOrderSn(userId int32) string {
	now := time.Now()
	return fmt.Sprintf("%d%d%d%d%d%d%d%d%d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), userId, rand.Intn(1000000))
}
