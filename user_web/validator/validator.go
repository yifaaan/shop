package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// 中国大陆手机号码正则：以 1 开头，第 2 位 3-9，后续 9 位数字，共 11 位
var mobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func ValidateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	return mobileRegex.MatchString(mobile)
}
