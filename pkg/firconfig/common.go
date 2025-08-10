package firconfig

import (
	"fmt"
	"os"
	"path/filepath"
)

// TLSConfig 定义TLS配置结构
type TLSConfig struct {
	CertFile   string `yaml:"certFile" env:"TLS_CERT_FILE"`
	KeyFile    string `yaml:"keyFile" env:"TLS_KEY_FILE"`
	CAFile     string `yaml:"caFile" env:"TLS_CA_FILE"`
	ClientAuth bool   `yaml:"clientAuth" env:"TLS_CLIENT_AUTH"`
	// 新增配置选项
	MinVersion         string `yaml:"minVersion" env:"TLS_MIN_VERSION"`
	MaxVersion         string `yaml:"maxVersion" env:"TLS_MAX_VERSION"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify" env:"TLS_INSECURE_SKIP_VERIFY"`

	// 内部字段，用于测试
	skipFileCheck bool
}

// DefaultTLSConfig 返回默认TLS配置
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		CertFile:           "certs/ca.crt",
		KeyFile:            "certs/ca.key",
		CAFile:             "certs/ca.crt",
		ClientAuth:         true,
		MinVersion:         "1.2",
		MaxVersion:         "1.3",
		InsecureSkipVerify: false,
	}
}

// IsEnabled 检查TLS是否启用
func (t *TLSConfig) IsEnabled() bool {
	return t != nil && t.CertFile != "" && t.KeyFile != "" && t.CAFile != ""
}

// Validate 验证TLS配置
func (t *TLSConfig) Validate() error {
	if !t.IsEnabled() {
		return nil // TLS未启用，无需验证
	}

	// 如果不是测试模式，检查证书文件是否存在
	if !t.skipFileCheck {
		if err := checkFileExists(t.CertFile); err != nil {
			return fmt.Errorf("certificate file: %w", err)
		}

		if err := checkFileExists(t.KeyFile); err != nil {
			return fmt.Errorf("private key file: %w", err)
		}

		if err := checkFileExists(t.CAFile); err != nil {
			return fmt.Errorf("CA certificate file: %w", err)
		}
	}

	// 验证TLS版本
	if t.MinVersion != "" {
		if !isValidTLSVersion(t.MinVersion) {
			return fmt.Errorf("invalid min TLS version: %s", t.MinVersion)
		}
	}

	if t.MaxVersion != "" {
		if !isValidTLSVersion(t.MaxVersion) {
			return fmt.Errorf("invalid max TLS version: %s", t.MaxVersion)
		}
	}

	return nil
}

// SkipFileCheck 设置跳过文件检查（用于测试）
func (t *TLSConfig) SkipFileCheck() {
	t.skipFileCheck = true
}

// checkFileExists 检查文件是否存在
func checkFileExists(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	// 如果是相对路径，尝试在当前目录和常见证书目录查找
	if !filepath.IsAbs(filePath) {
		// 检查当前目录
		if _, err := os.Stat(filePath); err == nil {
			return nil
		}

		// 检查常见证书目录
		commonPaths := []string{
			"certs",
			"certificates",
			"/etc/ssl/certs",
			"/usr/local/share/ca-certificates",
		}

		for _, path := range commonPaths {
			fullPath := filepath.Join(path, filePath)
			if _, err := os.Stat(fullPath); err == nil {
				return nil
			}
		}

		return fmt.Errorf("file not found: %s", filePath)
	}

	// 绝对路径
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	return nil
}

// isValidTLSVersion 验证TLS版本格式
func isValidTLSVersion(version string) bool {
	validVersions := map[string]bool{
		"1.0": true,
		"1.1": true,
		"1.2": true,
		"1.3": true,
	}
	return validVersions[version]
}
