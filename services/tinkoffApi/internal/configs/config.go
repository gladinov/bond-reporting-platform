package configs

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/russianinvestments/invest-api-go-sdk/investgo"
)

type Configs struct {
	TinkoffApiConfig *investgo.Config
	Config           *Config
}

type Config struct {
	Env               string          `env:"ENV" env-required:"true"`
	RootPath          string          `env:"ROOT_PATH" env-required:"true"`
	ConfigPath        string          `env:"CONFIG_PATH" env-required:"true"`
	Key               string          `env:"KEY" env-required:"true"`
	TinkoffApiAppPort string          `env:"TINKOFF_API_PORT" env-required:"true"`
	TinkoffApiAppHost string          `yaml:"TinkoffApiAppHost"`
	RedisHTTPServer   RedisHTTPServer `yaml:"redisHTTP"`
	Timeouts          Timeouts        `yaml:"timeouts"`
}

func (c *Config) GetTinkoffAppAddress() string {
	return getAddress(c.TinkoffApiAppHost, c.TinkoffApiAppPort)
}

type Timeouts struct {
	HTTPReadHeaderTimeout time.Duration `yaml:"http_read_header_timeout"`
	HTTPReadTimeout       time.Duration `yaml:"http_read_timeout"`
	HTTPWriteTimeout      time.Duration `yaml:"http_write_timeout"`
	HTTPIdleTimeout       time.Duration `yaml:"http_idle_timeout"`
	HTTPShutdownTimeout   time.Duration `yaml:"http_shutdown_timeout"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	AppCloseTimeout       time.Duration `yaml:"app_close_timeout"`
}

type RedisHTTPServer struct {
	Host     string `env:"REDIS_HOST" env-required:"true"`
	Port     string `env:"REDIS_PORT" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-required:"true"`
	// User        string
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

func (r *RedisHTTPServer) GetAddress() string {
	return getAddress(r.Host, r.Port)
}

func NewConfigs() *Configs {
	return &Configs{
		Config: &Config{
			RedisHTTPServer: RedisHTTPServer{},
		},
	}
}

func MustTinkoffConfigLoad(rootPath, configPath string) *investgo.Config {
	const op = "configs.MustTinkoffConfigLoad"
	Path := filepath.Join(rootPath, configPath)

	config, err := investgo.LoadConfig(Path)
	if err != nil {
		log.Fatalf("op: %s , can't load config", op)
	}

	return &config
}

func MustConfigLoad(rootPath, configPath string) *Config {
	const op = "configs.MustConfigLoad"
	Path := filepath.Join(rootPath, configPath)
	var cfg Config

	if err := cleanenv.ReadConfig(Path, &cfg); err != nil {
		log.Fatalf("%s:cannot read config:%s", op, err)
	}
	return &cfg
}

func MustInitConfigs() *Configs {
	configs := NewConfigs()
	envs, err := InjectEnvs()
	if err != nil {
		panic(err)
	}

	configs.Config = MustConfigLoad(envs.RootPath, envs.ConfigPath)
	configs.TinkoffApiConfig = MustTinkoffConfigLoad(envs.RootPath, envs.TinkoffConfigPath)
	return configs
}

type Envs struct {
	RootPath          string
	ConfigPath        string
	TinkoffConfigPath string
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

	tinkoffConfigPath := os.Getenv("TINKOFF_CONFIG_PATH")
	if tinkoffConfigPath == "" {
		return Envs{}, errors.New("TINKOFF_CONFIG_PATH environment variable is required")
	}

	envs := Envs{
		RootPath:          rootPath,
		ConfigPath:        configPath,
		TinkoffConfigPath: tinkoffConfigPath,
	}

	return envs, nil
}

func getAddress(host string, port string) string {
	return host + ":" + port
}
