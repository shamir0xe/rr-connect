package dependencies

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func NewViperConfig() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigFile("config.yaml")

	exe, err := os.Executable()
	if err == nil {
		cfg.AddConfigPath(filepath.Dir(exe))
	}
	cfg.AddConfigPath(".")

	cfg.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	cfg.AutomaticEnv()

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	log.Printf("✓ config loaded from: %s", cfg.ConfigFileUsed())
	log.Printf("✓ all keys: %v", cfg.AllKeys())
	return cfg, nil
}
