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
	Env                       string       `env:"ENV" env-required:"true"`
	RootPath                  string       `env:"ROOT_PATH" env-required:"true"`
	ConfigPath                string       `env:"CONFIG_PATH" env-required:"true"`
	SberConfigPath            string       `env:"SBER_CONFIG_PATH" env-required:"true"`
	DbType                    string       `yaml:"dbType"`
	ServiceStorageSQLLitePath string       `yaml:"serviceStorageSQLLitePath"`
	WorkersNubmer             int          `yaml:"workersNumber"`
	Clients                   Clients      `yaml:"clients"`
	PostgresHost              PostgresHost `yaml:"postgresHost"`
	Kafka                     Kafka
	Timeouts                  Timeouts `yaml:"timeouts"`
}

type Timeouts struct {
	HTTPReadHeaderTimeout time.Duration `yaml:"http_read_header_timeout"`
	HTTPReadTimeout       time.Duration `yaml:"http_read_timeout"`
	HTTPWriteTimeout      time.Duration `yaml:"http_write_timeout"`
	HTTPIdleTimeout       time.Duration `yaml:"http_idle_timeout"`
	HTTPShutdownTimeout   time.Duration `yaml:"http_shutdown_timeout"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	AppCloseTimeout       time.Duration `yaml:"app_close_timeout"`
	DBConnectTimeout      time.Duration `yaml:"db_connect_timeout"`
}

type Clients struct {
	TinkoffClient     TinkoffApiApp        `yaml:"tinkoffClient"`
	CBRClient         CBRApp               `yaml:"cbrClient"`
	MoexClient        MoexApp              `yaml:"moexClient"`
	BondReportService BondReportServiceApp `yaml:"bondReportService"`
}

type TinkoffApiApp struct {
	Host string `env:"TINKOFF_API_HOST" env-required:"true"`
	Port string `env:"TINKOFF_API_PORT" env-required:"true"`
}

func (t *TinkoffApiApp) GetTinkoffApiAddress() string {
	return getAddress(t.Host, t.Port)
}

type CBRApp struct {
	Host string `env:"CBR_HOST" env-required:"true"`
	Port string `env:"CBR_PORT" env-required:"true"`
}

func (c *CBRApp) GetCBRAppAddress() string {
	return getAddress(c.Host, c.Port)
}

type MoexApp struct {
	Host string `env:"MOEX_API_HOST" env-required:"true"`
	Port string `env:"MOEX_API_PORT" env-required:"true"`
}

func (m *MoexApp) GetMoexAppAddress() string {
	return getAddress(m.Host, m.Port)
}

type BondReportServiceApp struct {
	Host string `yaml:"bondReportServiceHost"`
	Port string `env:"BOND_REPORT_SERVICE_PORT" env-required:"true"`
}

func (b *BondReportServiceApp) GetBondReportServiceAppAddress() string {
	return getAddress(b.Host, b.Port)
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
	Port     string `env:"POSTGRES_SERVICE_PORT" env-required:"true"`
	PgUser   string `env:"PGUSER" env-required:"true"`
	SslMode  string `yaml:"sslmode" env:"POSTGRES_SSLMODE" env-default:"disable"`
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

func (p *PostgresHost) GetDSN() (string, error) {
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
	host := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=%s",
		p.User,
		p.Password,
		p.Host,
		p.Port,
		p.Dbname,
		p.SslMode)
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

func getAddress(host string, port string) string {
	return host + ":" + port
}

type Envs struct {
	RootPath       string
	ConfigPath     string
	SberConfigPath string
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

	sberConfigPath := os.Getenv("SBER_CONFIG_PATH")
	if sberConfigPath == "" {
		return Envs{}, errors.New("SBER_CONFIG_PATH environment variable is required")
	}

	envs := Envs{
		RootPath:       rootPath,
		ConfigPath:     configPath,
		SberConfigPath: sberConfigPath,
	}

	return envs, nil
}
