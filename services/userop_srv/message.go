package main

import (
	basemodel "shop/pkg/model"
)

// Message 留言：type 1留言 2投诉 3询问 4售后 5求购。
type Message struct {
	basemodel.BaseModel
	UserID  int32  `gorm:"index;not null"`
	Subject string `gorm:"type:varchar(100);not null"`
	Message string `gorm:"type:varchar(500);not null"`
	Type    int32  `gorm:"type:int;not null"`
	File    string `gorm:"type:varchar(200)"`
}