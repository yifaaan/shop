// Package nacosconf fetches service configuration from a Nacos config server
// and hands it to the caller as a populated *viper.Viper, with SHOP_-prefixed
// environment overrides applied on top (so secrets stay in the environment,
// not in Nacos).
package nacosconf

import (
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

// Options describes how to reach Nacos and which config entry to fetch.
type Options struct {
	Host      string // Nacos server host (HTTP/console port; SDK derives gRPC port = Port+1000)
	Port      int    // Nacos server HTTP port, e.g. 8848
	Namespace string // "" for the default "public" namespace
	Username  string // Nacos auth username (required when server auth is enabled)
	Password  string // Nacos auth password
	DataID    string // Nacos config Data ID
	Group     string // Nacos config Group
}

// Load connects to Nacos, fetches the YAML config at (DataID, Group), parses
// it into a viper instance with SHOP_-prefixed env overrides applied, and
// returns the viper. If onChange is non-nil, the config is also watched and
// onChange is invoked with a fresh viper on every Nacos-side change.
func Load(opts Options, onChange func(*viper.Viper)) (*viper.Viper, error) {
	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig: constant.NewClientConfig(
			constant.WithNamespaceId(opts.Namespace),
			constant.WithUsername(opts.Username),
			constant.WithPassword(opts.Password),
			constant.WithTimeoutMs(5000),
			constant.WithNotLoadCacheAtStart(true),
		),
		ServerConfigs: []constant.ServerConfig{
			*constant.NewServerConfig(opts.Host, uint64(opts.Port)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create nacos client: %w", err)
	}

	content, err := client.GetConfig(vo.ConfigParam{DataId: opts.DataID, Group: opts.Group})
	if err != nil {
		return nil, fmt.Errorf("get nacos config %s/%s: %w", opts.Group, opts.DataID, err)
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("nacos config %s/%s is empty", opts.Group, opts.DataID)
	}

	v, err := parse(content)
	if err != nil {
		return nil, fmt.Errorf("parse nacos config %s/%s: %w", opts.Group, opts.DataID, err)
	}

	if onChange != nil {
		if err := client.ListenConfig(vo.ConfigParam{
			DataId: opts.DataID,
			Group:  opts.Group,
			OnChange: func(_, _, _, data string) {
				if nv, err := parse(data); err == nil {
					onChange(nv)
				}
			},
		}); err != nil {
			return nil, fmt.Errorf("listen nacos config %s/%s: %w", opts.Group, opts.DataID, err)
		}
	}
	return v, nil
}

// parse builds a viper with SHOP_ env overrides and reads the YAML content.
func parse(content string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("SHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, err
	}
	return v, nil
}