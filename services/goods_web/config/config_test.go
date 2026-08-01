package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestUnmarshalConfigTraceDefaults(t *testing.T) {
	cfg := &Config{}
	if err := unmarshalConfig(viper.New(), cfg); err != nil {
		t.Fatalf("unmarshalConfig() error = %v", err)
	}

	if !cfg.Trace.Enabled {
		t.Fatal("Trace.Enabled = false, want true")
	}
	if cfg.Trace.Endpoint != "127.0.0.1:4317" {
		t.Fatalf("Trace.Endpoint = %q, want %q", cfg.Trace.Endpoint, "127.0.0.1:4317")
	}
	if !cfg.Trace.Insecure {
		t.Fatal("Trace.Insecure = false, want true")
	}
	if cfg.Trace.SampleRatio != 1 {
		t.Fatalf("Trace.SampleRatio = %v, want 1", cfg.Trace.SampleRatio)
	}
	if cfg.RateLimit.GoodsListQPS != 10 {
		t.Fatalf("RateLimit.GoodsListQPS = %v, want 10", cfg.RateLimit.GoodsListQPS)
	}
}

func TestUnmarshalConfigTraceEnvOverrides(t *testing.T) {
	t.Setenv("SHOP_TRACE_ENDPOINT", "jaeger:4317")
	t.Setenv("SHOP_TRACE_INSECURE", "false")
	t.Setenv("SHOP_TRACE_SAMPLE_RATIO", "0.25")
	t.Setenv("SHOP_RATE_LIMIT_GOODS_LIST_QPS", "2")

	v := viper.New()
	v.SetEnvPrefix("SHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	cfg := &Config{}
	if err := unmarshalConfig(v, cfg); err != nil {
		t.Fatalf("unmarshalConfig() error = %v", err)
	}

	if cfg.Trace.Endpoint != "jaeger:4317" {
		t.Fatalf("Trace.Endpoint = %q, want %q", cfg.Trace.Endpoint, "jaeger:4317")
	}
	if cfg.Trace.Insecure {
		t.Fatal("Trace.Insecure = true, want false")
	}
	if cfg.Trace.SampleRatio != 0.25 {
		t.Fatalf("Trace.SampleRatio = %v, want 0.25", cfg.Trace.SampleRatio)
	}
	if cfg.RateLimit.GoodsListQPS != 2 {
		t.Fatalf("RateLimit.GoodsListQPS = %v, want 2", cfg.RateLimit.GoodsListQPS)
	}
}
