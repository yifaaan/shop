package initialize

import (
	"fmt"
	"reflect"
	"shop/services/user_web/global"
	customValidator "shop/services/user_web/validator"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

func InitTranslator(locale string) (err error) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		zhT := zh.New()
		enT := en.New()
		uni := ut.New(enT, zhT, enT)
		global.Translator, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		switch locale {
		case "zh":
			_ = zhTranslations.RegisterDefaultTranslations(v, global.Translator)
		case "en":
			_ = enTranslations.RegisterDefaultTranslations(v, global.Translator)
		default:
			_ = zhTranslations.RegisterDefaultTranslations(v, global.Translator)
		}

		customValidator.Init()
		return nil
	}
	return nil
}