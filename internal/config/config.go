package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string              `env:"APP_ENV" env-default:"local"`
	HTTPServer    HTTPServerConfig    `env-prefix:"SERVER_" env-required:"true"`
	DB            DBConfig            `env-prefix:"DB_" env-required:"true"`
	Redis         RedisConfig         `env-prefix:"REDIS_" env-required:"true"`
	Auth          AuthConfig          `env-prefix:"AUTH_" env-required:"true"`
	LyricsAPI     LyricsAPIConfig     `env-prefix:"LYRICS_API_" env-required:"true"`
	TranslatorAPI TranslatorAPIConfig `env-prefix:"TRANSLATOR_API_" env-required:"true"`
}

type HTTPServerConfig struct {
	Host        string        `env:"HOST" env-required:"true"`
	Port        string        `env:"PORT" env-default:"8080"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type DBConfig struct {
	Host       string `env:"HOST" env-default:"localhost"`
	Port       string `env:"PORT" env-default:"5432"`
	DockerPort string `env:"DOCKER_PORT" env-default:"5432"`
	User       string `env:"USER" env-required:"true"`
	Password   string `env:"PASSWORD" env-required:"true"`
	Name       string `env:"NAME" env-required:"true"`
}

type RedisConfig struct {
	Host       string `env:"HOST" env-default:"localhost"`
	Port       string `env:"PORT" env-default:"6379"`
	DockerPort string `env:"DOCKER_PORT" env-default:"6379"`
	Password   string `env:"PASSWORD" env-required:"true"`
}

type AuthConfig struct {
	Host    string `env:"HOST" env-default:"localhost"`
	Port    string `env:"PORT" env-default:"44044"`
	Retries int    `env:"RETRIES" env-default:"5"`
}

type TranslatorAPIConfig struct {
	Key        string `env:"KEY" env-required:"true"`
	URL        string `env:"URL" env-required:"true"`
	TargetLang string `env:"TARGET_LANG" env-default:"ru"`
}

type LyricsAPIConfig struct {
	URL string `env:"URL" env-required:"true"`
}

// MustLoad Load config file and panic if error occurs
func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
