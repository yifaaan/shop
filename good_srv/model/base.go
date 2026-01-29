package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32          `gorm:"primaryKey;type:int" json:"id"`
	CreatedAt time.Time      `gorm:"column:add_time" json:"-"`
	IsDeleted int32          `gorm:"type:int;default:0;comment:'是否删除'" json:"-"`
	UpdatedAt time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type GormList []string

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}
