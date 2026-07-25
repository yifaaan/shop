package main

import (
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