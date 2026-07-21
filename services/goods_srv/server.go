package main

import (
	"shop/pkg/proto"

	"gorm.io/gorm"
)

// GoodsServer implements proto.GoodsServer over an injected *gorm.DB.
type GoodsServer struct {
	proto.UnimplementedGoodsServer
	db *gorm.DB
}

// NewGoodsServer wires a GoodsServer to its data store.
func NewGoodsServer(db *gorm.DB) *GoodsServer {
	return &GoodsServer{db: db}
}

// Paginate 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
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