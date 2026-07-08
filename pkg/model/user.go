package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32          `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"column:add_time"`
	UpdatedAt time.Time      `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	BaseModel
	Mobile   string     `gorm:"index:idx_mobile,unique;type:varchar(11);not null"`
	Password string     `gorm:"type:varchar(100);not null"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"` // 时间设置成指针，允许字段为null
	Gender   string     `gorm:"column:gender;default:male;type:varchar(6) comment 'female 女, male男'"`
	Role     int        `gorm:"column:role;default:1;type:int comment '1表示用户, 2表示管理员'"`
}
