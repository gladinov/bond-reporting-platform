package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"ENV" env-required:"true"`
	RootPath string `env:"ROOT_PATH" env-required:"true"`
	CertPath string `env:"CERT_PATH" env-required:"true"`
	KeyPath  string `env:"KEY_PATH" env-required:"true"`
	Clients  Clients
}
type Clients struct {
	TelegramMock TelegramMock
}

type TelegramMock struct {
	Token string `env:"LOCAL_BOT_TOKEN" env-required:"true"`
	Host  string `env:"TELEGRAM_MOCK_HOST" env-required:"true"`
	Port  string `env:"TELEGRAM_MOCK_PORT" env-required:"true"`
}

func (t *TelegramMock) GetTelegramMockAddress() string {
	return getAddress(t.Host, t.Port)
}

func getAddress(host string, port string) string {
	return host + ":" + port
}

func MustInitConfig() Config {
	const op = "config.MustInitConfig"

	var config Config
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return config
}
