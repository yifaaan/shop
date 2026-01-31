package main

import (
	"fmt"
	"os"
	"shop/user_web/global"
	"shop/user_web/initialize"
	"shop/user_web/utils"
	myValidators "shop/user_web/validator"

	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	initialize.InitLogger()
	// 初始化翻译器
	initialize.InitTrans("zh")
	// 初始化配置
	initialize.InitConfig()
	// 初始化rpc连接
	initialize.InitSrvConn()
	// 注册手机自定义验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("mobile", myValidators.ValidateMobile)
		_ = v.RegisterTranslation("mobile", global.Trans, func(ut ut.Translator) error {
			return ut.Add("mobile", "{0} 不是一个有效的手机号码", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			msg, err := ut.T("mobile", fe.Field())
			if err != nil {
				return fe.Field()
			}
			return msg
		})
	} else {
		panic("validator engine not found")
	}

	viper.AutomaticEnv()
	// debug时，port固定
	debug := viper.GetBool("SHOP_DEBUG")
	fmt.Println("SHOP_DEBUG env value:", os.Getenv("SHOP_DEBUG"))
	fmt.Println("debug ", debug)
	if !debug {
		port, err := utils.GetFreePort()
		if err != nil {
			zap.S().Fatalf("get free port failed: %v", err)
		}
		global.ServerConfig.Port = port
	}
	// 初始化路由
	r := initialize.Routers()
	zap.S().Infof("server run at port %s:%d", global.ServerConfig.IP, global.ServerConfig.Port)
	err := r.Run(fmt.Sprintf("%s:%d", global.ServerConfig.IP, global.ServerConfig.Port))
	if err != nil {
		zap.S().Errorf("server run failed: %v", err)
		os.Exit(1)
	}
}
