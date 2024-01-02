package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	UNIDOC_LICENSE_API_KEY string `mapstructure:"UNIDOC_LICENSE_API_KEY"`
	RedisTLS               string `mapstructure:"REDIS_TLS"`
	RedisAdrr              string `mapstructure:"REDIS_ADDR"`
	RedisUsername          string `mapstructure:"REDIS_USERNAME"`
	RedisPassword          string `mapstructure:"REDIS_PASSWORD"`
	OpenApiKey             string `mapstructure:"OPEN_API_KEY"`
	KakaTopic              string `mapstructure:"KAFKA_TOPIC"`
}

var Cfg *Config

func InitConfig(path string) {
	viper.SetConfigName("app.development")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		panic(err)
	}
}
