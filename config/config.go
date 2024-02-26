package config

import (
	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

type Config struct {
	MongoDBURI            string `env:"MONGODB_URI"`
	OrderCollection       string `env:"ORDER_COLLECTION"`
	OrderStatusCollection string `env:"ORDER_STATUS_COLLECTION"`
	RabbitMQURI           string `json:"RABBITMQ_URI"`
	OrderQueueName        string `json:"ORDER_QUEUE_NAME"`
	OrderErrorQueueName   string `json:"ORDER_ERROR_QUEUE_NAME"`
	OrderExchangeName     string `json:"ORDER_EXCHANGE_NAME"`
	OrderExchangeType     string `json:"ORDER_EXCHANGE_TYPE"`
}

func NewConfig() Config {
	godotenv.Load()
	config := Config{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}
	return config
}
