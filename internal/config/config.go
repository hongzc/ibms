package config

import "github.com/spf13/viper"

// Config 应用配置。
type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
}

// Server HTTP 服务配置。
type Server struct {
	Port int `mapstructure:"port"`
}

// Database 数据库配置。
type Database struct {
	DSN string `mapstructure:"dsn"`
}

// Load 从指定路径读取 YAML 配置文件。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
