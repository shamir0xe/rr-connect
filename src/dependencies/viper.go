package dependencies

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func NewViperConfig() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")

	exe, err := os.Executable()
	if err == nil {
		cfg.AddConfigPath(filepath.Dir(exe))
	}
	cfg.AddConfigPath(".")

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return cfg, nil
}
