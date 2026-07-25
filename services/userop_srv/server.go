package main

import (
	"strings"

	"shop/pkg/proto"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserOpServer 实现 proto.UserOpServer，持 *gorm.DB。
// 三个域（收藏/地址/留言）的 handler 分别在 userfav.go/address.go/message.go。
type UserOpServer struct {
	proto.UnimplementedUserOpServer
	db  *gorm.DB
	log *zap.SugaredLogger
}

// NewUserOpServer 装配 server。
func NewUserOpServer(db *gorm.DB, log *zap.SugaredLogger) *UserOpServer {
	return &UserOpServer{db: db, log: log}
}

// isDuplicateKeyErr 判定 gorm 错误是否为唯一索引冲突（MySQL error 1062）。
// 用于收藏等"重复视为已存在"的语义。
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "1062") || strings.Contains(msg, "Duplicate entry")
}