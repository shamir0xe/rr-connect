package dependencies

import "github.com/spf13/viper"

func NewViperConfig() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(".")
	cfg.ReadInConfig()
	return cfg, nil
}
