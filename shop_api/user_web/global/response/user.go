package response

import (
	"strconv"
	"time"
)

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	formatted := time.Time(j).Format("2006-01-02")
	// 添加引号
	quoted := strconv.Quote(formatted)
	return []byte(quoted), nil
}

// UserResponse 用户响应结构体
type UserResponse struct {
	Id       int32    `json:"id"`
	Role     int32    `json:"role"`
	Birthday JsonTime `json:"birthday"`
	NickName string   `json:"nick_name"`
	Mobile   string   `json:"mobile"`
	Gender   string   `json:"gender"`
}
