package configs

import (
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
	CbrHost    string   `yaml:"cbrHost"`
	Clients    Clients  `yaml:"clients"`
	Timeouts   Timeouts `yaml:"timeouts"`
}

type Clients struct {
	CbrAppApiClient CbrAppApiClient
}

type Timeouts struct {
	HTTPReadHeaderTimeout time.Duration `yaml:"http_read_header_timeout"`
	HTTPReadTimeout       time.Duration `yaml:"http_read_timeout"`
	HTTPWriteTimeout      time.Duration `yaml:"http_write_timeout"`
	HTTPIdleTimeout       time.Duration `yaml:"http_idle_timeout"`
	HTTPShutdownTimeout   time.Duration `yaml:"http_shutdown_timeout"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
}

type CbrAppApiClient struct {
	Host string `yaml:"cbr_app_host"`
	Port string `env:"CBR_PORT" env-required:"true"`
}

func (c *CbrAppApiClient) GetCbrAppServer() string {
	return getAddress(c.Host, c.Port)
}

func MustInitConfig() Config {
	const op = "configs.MustInitConfig"
	envs, err := InjectEnvs()
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(envs.RootPath, envs.ConfigPath)

	var cfg Config
	if err = cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("%s:cannot read config:%s", op, err)
	}

	return cfg
}

type Envs struct {
	RootPath   string
	ConfigPath string
}

func InjectEnvs() (Envs, error) {
	rootPath := os.Getenv("ROOT_PATH")
	if rootPath == "" {
		panic("ROOT_PATH environment variable is required")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("CONFIG_PATH environment variable is required")
	}
	var envs Envs
	envs.RootPath = rootPath
	envs.ConfigPath = configPath

	return envs, nil
}

func getAddress(host string, port string) string {
	return host + ":" + port
}
