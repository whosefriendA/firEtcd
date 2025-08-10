package firconfig

import (
	"testing"
)

func TestTLSConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *TLSConfig
		wantErr bool
	}{
		{
			name:    "nil TLS config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "empty cert file",
			config: &TLSConfig{
				CertFile: "",
				KeyFile:  "certs/server.key",
				CAFile:   "certs/ca.crt",
			},
			wantErr: false, // 空配置不算错误，只是未启用
		},
		{
			name: "invalid min version",
			config: &TLSConfig{
				CertFile:   "certs/server.crt",
				KeyFile:    "certs/server.key",
				CAFile:     "certs/ca.crt",
				MinVersion: "2.0", // 无效版本
			},
			wantErr: true,
		},
		{
			name: "invalid max version",
			config: &TLSConfig{
				CertFile:   "certs/server.crt",
				KeyFile:    "certs/server.key",
				CAFile:     "certs/ca.crt",
				MaxVersion: "2.0", // 无效版本
			},
			wantErr: true,
		},
		{
			name: "valid version range",
			config: &TLSConfig{
				CertFile:   "certs/server.crt",
				KeyFile:    "certs/server.key",
				CAFile:     "certs/ca.crt",
				MinVersion: "1.2",
				MaxVersion: "1.3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 跳过文件检查，只测试配置验证逻辑
			if tt.config != nil {
				tt.config.SkipFileCheck()
			}

			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TLSConfig.Validate() error = %v, wantErr %v", err, tt.name)
			}
		})
	}
}

func TestGatewayConfigValidation(t *testing.T) {
	// 测试无TLS配置的情况
	gateway := &Gateway{
		Addr:    "127.0.0.1",
		Port:    ":51030",
		BaseUrl: "/firEtcd",
		TLS:     nil, // 无TLS配置
	}

	err := gateway.Validate()
	if err != nil {
		t.Errorf("Gateway.Validate() error = %v", err)
	}
}

func TestKvserverConfigValidation(t *testing.T) {
	// 测试无TLS配置的情况
	kvserver := &Kvserver{
		Addr:         "127.0.0.1",
		Port:         ":51242",
		DataBasePath: "data",
		Maxraftstate: 100000,
		TLS:          nil, // 无TLS配置
	}

	err := kvserver.Validate()
	if err != nil {
		t.Errorf("Kvserver.Validate() error = %v", err)
	}
}

func TestTLSConfigIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *TLSConfig
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "empty config",
			config:   &TLSConfig{},
			expected: false,
		},
		{
			name: "partial config",
			config: &TLSConfig{
				CertFile: "certs/server.crt",
				KeyFile:  "certs/server.key",
				// 缺少CAFile
			},
			expected: false,
		},
		{
			name: "complete config",
			config: &TLSConfig{
				CertFile: "certs/server.crt",
				KeyFile:  "certs/server.key",
				CAFile:   "certs/ca.crt",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsEnabled()
			if result != tt.expected {
				t.Errorf("TLSConfig.IsEnabled() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultTLSConfig(t *testing.T) {
	config := DefaultTLSConfig()

	if config.CertFile != "certs/ca.crt" {
		t.Errorf("DefaultTLSConfig.CertFile = %v, expected certs/ca.crt", config.CertFile)
	}

	if config.KeyFile != "certs/ca.key" {
		t.Errorf("DefaultTLSConfig.KeyFile = %v, expected certs/ca.key", config.KeyFile)
	}

	if config.CAFile != "certs/ca.crt" {
		t.Errorf("DefaultTLSConfig.CAFile = %v, expected certs/ca.crt", config.CAFile)
	}

	if !config.ClientAuth {
		t.Errorf("DefaultTLSConfig.ClientAuth = %v, expected true", config.ClientAuth)
	}

	if config.MinVersion != "1.2" {
		t.Errorf("DefaultTLSConfig.MinVersion = %v, expected 1.2", config.MinVersion)
	}

	if config.MaxVersion != "1.3" {
		t.Errorf("DefaultTLSConfig.MaxVersion = %v, expected 1.3", config.MaxVersion)
	}

	if config.InsecureSkipVerify {
		t.Errorf("DefaultTLSConfig.InsecureSkipVerify = %v, expected false", config.InsecureSkipVerify)
	}
}

func TestDefaultGateway(t *testing.T) {
	gateway := DefaultGateway()

	if gateway.Addr != "127.0.0.1" {
		t.Errorf("DefaultGateway.Addr = %v, expected 127.0.0.1", gateway.Addr)
	}

	if gateway.Port != ":51030" {
		t.Errorf("DefaultGateway.Port = %v, expected :51030", gateway.Port)
	}

	if gateway.BaseUrl != "/firEtcd" {
		t.Errorf("DefaultGateway.BaseUrl = %v, expected /firEtcd", gateway.BaseUrl)
	}

	if gateway.TLS == nil {
		t.Error("DefaultGateway.TLS should not be nil")
	}
}

func TestDefaultKVServer(t *testing.T) {
	kvserver := DefaultKVServer()

	if kvserver.Addr != "127.0.0.1" {
		t.Errorf("DefaultKVServer.Addr = %v, expected 127.0.0.1", kvserver.Addr)
	}

	if kvserver.Port != ":51242" {
		t.Errorf("DefaultKVServer.Port = %v, expected :51242", kvserver.Port)
	}

	if kvserver.DataBasePath != "data" {
		t.Errorf("DefaultKVServer.DataBasePath = %v, expected data", kvserver.DataBasePath)
	}

	if kvserver.Maxraftstate != 100000 {
		t.Errorf("DefaultKVServer.Maxraftstate = %v, expected 100000", kvserver.Maxraftstate)
	}

	if kvserver.TLS == nil {
		t.Error("DefaultKVServer.TLS should not be nil")
	}
}
