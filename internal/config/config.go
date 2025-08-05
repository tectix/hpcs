package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Cache   CacheConfig   `mapstructure:"cache"`
	Cluster ClusterConfig `mapstructure:"cluster"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	MaxConnections int           `mapstructure:"max_connections"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
}

type CacheConfig struct {
	MaxMemory       string        `mapstructure:"max_memory"`
	EvictionPolicy  string        `mapstructure:"eviction_policy"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type ClusterConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	Nodes        []string `mapstructure:"nodes"`
	ReplicaCount int      `mapstructure:"replica_count"`
	VirtualNodes int      `mapstructure:"virtual_nodes"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
	File   string `mapstructure:"file"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	
	setDefaults()
	
	viper.AutomaticEnv()
	viper.SetEnvPrefix("HPCS")
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 6379)
	viper.SetDefault("server.max_connections", 10000)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	
	viper.SetDefault("cache.max_memory", "1GB")
	viper.SetDefault("cache.eviction_policy", "lru")
	viper.SetDefault("cache.cleanup_interval", "60s")
	
	viper.SetDefault("cluster.enabled", false)
	viper.SetDefault("cluster.replica_count", 1)
	viper.SetDefault("cluster.virtual_nodes", 150)
	
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", 8080)
	viper.SetDefault("metrics.path", "/metrics")
	
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
}

func validate(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	
	if config.Server.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}
	
	validPolicies := map[string]bool{"lru": true, "lfu": true, "random": true}
	if !validPolicies[config.Cache.EvictionPolicy] {
		return fmt.Errorf("invalid eviction policy: %s", config.Cache.EvictionPolicy)
	}
	
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}
	
	return nil
}