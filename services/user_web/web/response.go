package web

import (
	"fmt"
	"time"
)

// JsonTime is a time.Time that marshals to a "YYYY-MM-DD" JSON string.
type JsonTime time.Time

func (t JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02"))), nil
}

type UserResponse struct {
	Id       int32    `json:"id"`
	NickName string   `json:"name"`
	Mobile   string   `json:"mobile"`
	Birthday JsonTime `json:"birthday"`
	Gender   string   `json:"gender"`
}
