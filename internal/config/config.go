package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"development"`
	Server   ServerConfig   `yaml:"http_server"`
	DB       DBConfig       `yaml:"database"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Cache    CacheConfig    `yaml:"cache"`
	Producer ProducerConfig `yaml:"producer"`
}

type ServerConfig struct {
	Host string `yaml:"host" env-required:"true"`
	Port string `yaml:"port" env-required:"true"`
}

type DBConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Name     string `yaml:"name" env-required:"true"`
	SslMode  string `yaml:"sslmode" env-default:"disable"`
}

type CacheConfig struct {
	Size     int `yaml:"size" env-default:"0"`
	Capacity int `yaml:"capacity" env-default:"10"`
}

type ProducerConfig struct {
	NmbOfOrders int `yaml:"nmb_of_orders" env-default:"3"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env-required:"true"`
	Topic   string   `yaml:"topic" env-required:"true"`
	GroupID string   `yaml:"group_id" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file not found at %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return &cfg
}
