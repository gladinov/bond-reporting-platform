package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string `env:"ENV" env-required:"true"`
	Clients Clients
	Kafka   Kafka
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

type Clients struct {
	NotificationService NotificationService
	Telegram            Telegram
}

type Telegram struct {
	Token string `env:"LOCAL_BOT_TOKEN" env-required:"true"`
	Host  string `env:"TELEGRAM_HOST" env-required:"true"`
}
type NotificationService struct {
	Host string `env:"NOTIFICATION_SERVICE_HOST" env-required:"true"`
	Port string `env:"NOTIFICATION_SERVICE_PORT" env-required:"true"`
}

func (n *NotificationService) GetNotificationServiceAddress() string {
	return getAddress(n.Host, n.Port)
}

type Kafka struct {
	Host string `env:"KAFKA_HOST" env-required:"true"`
	Port string `env:"KAFKA_PORT" env-required:"true"`
}

func (k *Kafka) GetKafkaAddress() string {
	return getAddress(k.Host, k.Port)
}

func getAddress(host string, port string) string {
	return host + ":" + port
}
