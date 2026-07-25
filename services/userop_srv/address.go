package main

import (
	basemodel "shop/pkg/model"
)

// Address 收货地址：按 user_id 归属。
type Address struct {
	basemodel.BaseModel
	UserID       int32  `gorm:"index;not null"`
	Province     string `gorm:"type:varchar(20);not null"`
	City         string `gorm:"type:varchar(20);not null"`
	District     string `gorm:"type:varchar(20);not null"`
	Address      string `gorm:"type:varchar(100);not null"`
	SignerName   string `gorm:"type:varchar(30);not null"`
	SignerMobile string `gorm:"type:varchar(20);not null"`
}