package main

import (
	basemodel "shop/pkg/model"
)

// UserFav 收藏：(user_id, goods_id) 唯一，防重复收藏。
type UserFav struct {
	basemodel.BaseModel
	UserID  int32 `gorm:"index:idx_user_goods,unique;not null"`
	GoodsID int32 `gorm:"index:idx_user_goods,unique;not null"`
}