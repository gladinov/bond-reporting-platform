package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                string          `env:"ENV" env-required:"true"`
	RootPath           string          `env:"ROOT_PATH" env-required:"true"`
	ConfigPath         string          `env:"CONFIG_PATH" env-required:"true"`
	Key                string          `env:"KEY" env-required:"true"`
	Token              string          `env:"LOCAL_BOT_TOKEN" env-required:"true"`
	ClientsHosts       Clients         `yaml:"clients"`
	DbType             string          `yaml:"dbType"`
	StorageSQLLitePath string          `yaml:"storageSQLLitePath"`
	PostgresHost       PostgresHost    `yaml:"postgresHost"`
	RedisHTTPServer    RedisHTTPServer `yaml:"redis"`
	Kafka              Kafka
}

type Kafka struct {
	Host string `env:"KAFKA_HOST" env-required:"true"`
	Port string `env:"KAFKA_PORT" env-required:"true"`
}

func (k *Kafka) GetKafkaAddress() string {
	return getAddress(k.Host, k.Port)
}

type PostgresHost struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Dbname   string `env:"POSTGRES_DB" env-required:"true"`
	Port     string `env:"POSTGRES_USER_PORT" env-required:"true"`
	PgUser   string `env:"PGUSER" env-required:"true"`
	SslMode  string `yaml:"sslmode"`
}

type Clients struct {
	BondReportServiceHost string `env:"BOND_REPORT_SERVICE_HOST" env-required:"true"`
	BondReportServicePort string `env:"BOND_REPORT_SERVICE_PORT" env-required:"true"`
	TinkoffApiHost        string `env:"TINKOFF_API_HOST" env-required:"true"`
	TinkoffApiPort        string `env:"TINKOFF_API_PORT" env-required:"true"`
	TelegramHost          string `env:"TELEGRAM_HOST" env-required:"true"`
}

func (r *Clients) GetTinkoffApiAddress() string {
	return getAddress(r.TinkoffApiHost, r.TinkoffApiPort)
}

func (r *Clients) GetBondReportAddress() string {
	return getAddress(r.BondReportServiceHost, r.BondReportServicePort)
}

type RedisHTTPServer struct {
	Host        string        `env:"REDIS_HOST" env-required:"true"`
	Port        string        `env:"REDIS_PORT" env-required:"true"`
	Password    string        `env:"REDIS_PASSWORD" env-required:"true"`
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

func (r *RedisHTTPServer) GetAddress() string {
	return getAddress(r.Host, r.Port)
}

func getAddress(host string, port string) string {
	return host + ":" + port
}

func (p *PostgresHost) GetStringHost() (string, error) {
	if p.Host == "" {
		return "", errors.New("empty host in config")
	}
	if p.User == "" {
		return "", errors.New("empty user in config")
	}
	if p.Password == "" {
		return "", errors.New("empty password in config")
	}
	if p.Dbname == "" {
		return "", errors.New("empty dbname in config")
	}
	if p.Port == "" {
		return "", errors.New("empty port in config")
	}
	host := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=%s",
		p.Host,
		p.User,
		p.Password,
		p.Dbname,
		p.Port,
		p.SslMode)
	return host, nil
}

func (p *PostgresHost) GetAdress() string {
	return getAddress(p.Host, p.Port)
}

func (p *PostgresHost) GetHostToGoMigrate() (string, error) {
	if p.Host == "" {
		return "", errors.New("empty host in config")
	}
	if p.User == "" {
		return "", errors.New("empty user in config")
	}
	if p.Password == "" {
		return "", errors.New("empty password in config")
	}
	if p.Dbname == "" {
		return "", errors.New("empty dbname in config")
	}
	if p.Port == "" {
		return "", errors.New("empty port in config")
	}
	host := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=%s", p.User, p.Password, p.Host, p.Port, p.Dbname, p.SslMode)
	return host, nil
}

func MustInitConfig() Config {
	const op = "config.MustInitConfig"
	envs, err := InjectEnvs()
	if err != nil {
		log.Fatalf("%s: %s", op, err)
	}
	Path := filepath.Join(envs.RootPath, envs.ConfigPath)

	var config Config
	err = cleanenv.ReadConfig(Path, &config)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return config
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

	envs := Envs{
		RootPath:   rootPath,
		ConfigPath: configPath,
	}

	return envs, nil
}
