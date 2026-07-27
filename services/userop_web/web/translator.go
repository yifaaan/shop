package web

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// translator is set once by ConfigureTranslator at startup and then read
// concurrently by request handlers.
var translator ut.Translator

// ConfigureTranslator wires the gin binding validator with zh/en
// translations and the custom "mobile" rule. It must be called once from
// main before serving traffic.
func ConfigureTranslator(locale string) error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil
	}

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
	var got bool
	translator, got = uni.GetTranslator(locale)
	if !got {
		return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
	}

	switch locale {
	case "zh":
		_ = zhTranslations.RegisterDefaultTranslations(v, translator)
	case "en":
		_ = enTranslations.RegisterDefaultTranslations(v, translator)
	default:
		_ = zhTranslations.RegisterDefaultTranslations(v, translator)
	}

	registerMobileValidator(v)
	return nil
}

func registerMobileValidator(v *validator.Validate) {
	_ = v.RegisterValidation("mobile", validateMobile)
	_ = v.RegisterTranslation("mobile", translator, func(ut ut.Translator) error {
		return ut.Add("mobile", "{0} 非法的手机号码!", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})
}

func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, mobile)
	return matched
}

func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}