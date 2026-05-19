package setting

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const configEnvKey = "APP_CONFIG"

// Setting wraps viper for reading application configuration sections.
type Setting struct {
	vp *viper.Viper
}

// NewSetting initializes configuration loading.
//
// Config file priority:
//  1. first non-empty configFiles argument
//  2. APP_CONFIG environment variable
//  3. configs/config.yaml
func NewSetting(configFiles ...string) (*Setting, error) {
	vp := viper.New()

	configFile := firstNonEmpty(configFiles...)
	if configFile == "" {
		configFile = strings.TrimSpace(os.Getenv(configEnvKey))
	}
	configureConfigFile(vp, configFile)

	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	return &Setting{vp: vp}, nil
}

func configureConfigFile(vp *viper.Viper, configFile string) {
	if configFile == "" {
		vp.SetConfigName("config")
		vp.AddConfigPath("configs")
		vp.SetConfigType("yaml")
		return
	}

	if filepath.IsAbs(configFile) || filepath.Dir(configFile) != "." {
		vp.SetConfigFile(configFile)
		return
	}

	ext := filepath.Ext(configFile)
	vp.SetConfigName(strings.TrimSuffix(configFile, ext))
	vp.AddConfigPath(".")
	vp.AddConfigPath("configs")
	if ext == "" {
		vp.SetConfigType("yaml")
		return
	}
	vp.SetConfigType(strings.TrimPrefix(ext, "."))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
