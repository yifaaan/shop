package main

import (
	"shop/pkg/proto"
	"shop/services/goods_srv/esearch"

	"gorm.io/gorm"
)

// GoodsServer implements proto.GoodsServer over an injected *gorm.DB,
// 关键词检索走注入的 esearch.Service。
type GoodsServer struct {
	proto.UnimplementedGoodsServer
	db *gorm.DB
	es *esearch.Service
}

// NewGoodsServer wires a GoodsServer to its data store and ES search service。
func NewGoodsServer(db *gorm.DB, es *esearch.Service) *GoodsServer {
	return &GoodsServer{db: db, es: es}
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