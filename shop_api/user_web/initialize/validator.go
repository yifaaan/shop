package initialize

import (
	"reflect"
	"shop/shop_api/user_web/global"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

func InitTrans(locale string) {
	// 初始化翻译器
	global.Uni = ut.New(en.New(), zh.New())
	// 初始化验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// register function to get tag name from json tags.
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// 注册翻译器
		trans, ok := global.Uni.GetTranslator(locale)
		if !ok {
			trans = global.Uni.GetFallback()
		}
		switch locale {
		case "en":
			en_translations.RegisterDefaultTranslations(v, trans)
		case "zh":
			zh_translations.RegisterDefaultTranslations(v, trans)
		default:
			zh_translations.RegisterDefaultTranslations(v, trans)
		}
	}
}
