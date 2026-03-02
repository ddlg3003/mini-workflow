package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Frontend ServiceConfig  `mapstructure:"frontend"`
	History  ServiceConfig  `mapstructure:"history"`
	Matching ServiceConfig  `mapstructure:"matching"`
	Log      LogConfig      `mapstructure:"log"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
}

type ServiceConfig struct {
	GRPCPort             int    `mapstructure:"grpc_port"`
	Addr                 string `mapstructure:"addr"`
	ClientTimeoutSeconds int    `mapstructure:"client_timeout_seconds"`
	PollTimeoutSeconds   int    `mapstructure:"poll_timeout_seconds"`
}

func (s ServiceConfig) ListenAddr() string {
	return fmt.Sprintf(":%d", s.GRPCPort)
}

func (s ServiceConfig) ClientTimeout() time.Duration {
	return time.Duration(s.ClientTimeoutSeconds) * time.Second
}

func (s ServiceConfig) PollTimeout() time.Duration {
	return time.Duration(s.PollTimeoutSeconds) * time.Second
}

type LogConfig struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
}

func Load(paths ...string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	for _, p := range paths {
		v.AddConfigPath(p)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("..")
	v.AddConfigPath("../..")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
