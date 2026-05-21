package configs

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string   `env:"ENV" env-required:"true"`
	RootPath   string   `env:"ROOT_PATH" env-required:"true"`
	ConfigPath string   `env:"CONFIG_PATH" env-required:"true"`
	MoexHost   string   `yaml:"moexHost"`
	Clients    Clients  `yaml:"clients"`
	Timeouts   Timeouts `yaml:"timeouts"`
}

type Clients struct {
	MoexApiAppClient MoexApiAppClient
}

type Timeouts struct {
	HTTPReadHeaderTimeout time.Duration `yaml:"http_read_header_timeout"`
	HTTPReadTimeout       time.Duration `yaml:"http_read_timeout"`
	HTTPWriteTimeout      time.Duration `yaml:"http_write_timeout"`
	HTTPIdleTimeout       time.Duration `yaml:"http_idle_timeout"`
	HTTPShutdownTimeout   time.Duration `yaml:"http_shutdown_timeout"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
}

type MoexApiAppClient struct {
	Host string `yaml:"moexApiAppHost"`
	Port string `env:"MOEX_API_PORT" env-required:"true"`
}

func (c *MoexApiAppClient) GetMoexApiAppClientAddress() string {
	return getAddress(c.Host, c.Port)
}

func MustLoad() *Config {
	const op = "configs.mustLoad"
	envs, err := InjectEnvs()
	if err != nil {
		log.Fatalf("%s:%s", op, err)
	}
	configPath := filepath.Join(envs.RootPath, envs.ConfigPath)
	if configPath == "" {
		log.Fatalf("%s: CONFIG_PATH is not set", op)
	}

	var cfg Config

	if err = cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("%s:cannot read config:%s", op, err)
	}

	return &cfg
}

type Envs struct {
	RootPath   string
	ConfigPath string
}

func InjectEnvs() (Envs, error) {
	rootPath := os.Getenv("ROOT_PATH")
	if rootPath == "" {
		return Envs{}, errors.New("ROOT_PATH environment variable is required")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return Envs{}, errors.New("CONFIG_PATH environment variable is required")
	}

	envs := Envs{RootPath: rootPath,
		ConfigPath: configPath}

	return envs, nil
}

func getAddress(host string, port string) string {
	return host + ":" + port
}
