package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port    int    `mapstructure:"port"`
	SaveDir string `mapstructure:"saveDir"`
	NoSave  bool   `mapstructure:"noSave"`
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	localDir := homeDir + "/.local/share/gochat"
	if err = os.MkdirAll(localDir+"/data", 0755); err != nil {
		return nil, err
	}
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetDefault("port", 8080)
	viper.SetDefault("saveDir", localDir+"/data")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
