package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	env        string `yaml:"env" required:"true"`
	Datasource `yaml:"datasource" required:"true"`
	HttpServer `yaml:"http_server" required:"true"`
}

type Datasource struct {
	Host     string `yaml:"host" required:"true"`
	Port     int    `yaml:"port" default:"5432"`
	Database string `yaml:"database" required:"true"`
	User     string `yaml:"user" required:"true"`
	Pass     string `yaml:"pass" required:"true"`
}

type HttpServer struct {
	Host        string        `yaml:"host" required:"true"`
	Port        int           `yaml:"port" default:"8080"`
	Timeout     time.Duration `yaml:"timeout" default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"30s"`
}

func MustLoafConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("CONFIG_PATH does not exist")
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatal(err)
	}

	return &config
}
