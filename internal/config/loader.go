package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	PgSql struct {
		Host         string
		Port         int
		User         string
		Password     string
		DatabaseName string
	}
	SecretKey       string
	MiddlewaresKeys struct {
		Response struct {
			Response   string
			StatusCode string
			Message    string
		}
		Validate struct {
			Validated string
		}
		Auth struct {
			TokenKey string
		}
	}
	ConfigKey string
	Minio     struct {
	}
}

func LoadConfig() (*Config, error) {
	// 获取程序当前的工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %s", err)
	}
	configFilePath := filepath.Join(currentDir, "internal", "config", "config.yaml")
	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
