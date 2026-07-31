package telemetry

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid",
			cfg: Config{
				ServiceName: "order_web",
				Endpoint:    "127.0.0.1:4317",
				SampleRatio: 1,
			},
		},
		{
			name:    "missing service name",
			cfg:     Config{Endpoint: "127.0.0.1:4317", SampleRatio: 1},
			wantErr: true,
		},
		{
			name:    "missing endpoint",
			cfg:     Config{ServiceName: "order_web", SampleRatio: 1},
			wantErr: true,
		},
		{
			name:    "sample ratio too high",
			cfg:     Config{ServiceName: "order_web", Endpoint: "127.0.0.1:4317", SampleRatio: 1.1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
