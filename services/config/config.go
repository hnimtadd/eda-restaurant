package config

import (
	"log"

	"github.com/spf13/viper"
)

type MongoConfig struct {
	Source        string `mapstructure:"db_source"`
	Username      string `mapstructure:"db_username"`
	Password      string `mapstructure:"db_password"`
	Database      string `mapstructure:"db_database"`
	AuthSource    string `mapstructure:"db_authsource"`
	AuthMechanism string `mapstructure:"db_authmechanism"`
}

type RabbitmqConfig struct {
	Source   string `mapstructure:"rabbitmq_source"`
	Username string `mapstructure:"rabbitmq_username"`
	Password string `mapstructure:"rabbitmq_password"`
}

func LoadConfig(path string, v any) error {
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(v); err != nil {
		return err
	}
	return nil
}
func NewMongoConfig(path string) MongoConfig {
	var config MongoConfig
	if err := LoadConfig(path, &config); err != nil {
		log.Fatalf("error: %v", err)
	}
	return config
}

func NewRabbitmqConfig(path string) RabbitmqConfig {
	var config RabbitmqConfig
	if err := LoadConfig(path, &config); err != nil {
		log.Panicf("error: %v", err)
	}
	return config
}
