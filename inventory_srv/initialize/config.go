package initialize

import (
	"encoding/json"
	"fmt"
	"os"
	"shop/inventory_srv/global"
	"shop/inventory_srv/utils"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func getEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func InitConfig() {
	debug := false
	if getEnvInfo("SHOP_DEBUG") {
		debug = true
		viper.SetConfigFile("./inventory_srv/config-debug.yaml")
	} else {
		viper.SetConfigFile("./inventory_srv/config-pro.yaml")
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// 读取配置到全局变量
	err = viper.Unmarshal(&global.NacosConfig)
	if err != nil {
		panic(err)
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:   global.NacosConfig.Host,
			Port:     global.NacosConfig.Port,
			GrpcPort: global.NacosConfig.GrpcPort,
		},
	}
	logDir := "tmp/nacos/log"
	cacheDir := "tmp/nacos/cache"
	for _, dir := range []string{logDir, cacheDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			zap.S().Fatalf("create nacos dir failed: %v", err)
		}
	}

	clientConfig := constant.ClientConfig{
		NamespaceId:         global.NacosConfig.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              logDir,
		CacheDir:            cacheDir,
		LogLevel:            "debug",
	}

	// nacos配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		fmt.Printf("Failed to create Nacos config client: %v\n", err)
		os.Exit(1)
	}

	// 尝试获取 user-srv 配置
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.DataID,
		Group:  global.NacosConfig.Group,
	})
	if err != nil {
		zap.S().Fatalf("get nacos config failed: %v", err)
	}
	err = json.Unmarshal([]byte(content), &global.ServerConfig)
	if err != nil {
		zap.S().Fatalf("unmarshal nacos config failed: %v", err)
	}

	if !debug {
		var err error
		global.ServerConfig.Port, err = utils.GetFreePort()
		if err != nil {
			zap.S().Fatalf("get free port failed: %v", err)
		}
	}
	fmt.Println(global.ServerConfig)
}
