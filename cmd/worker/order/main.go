package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/panggggg/order-service/config"
	"github.com/panggggg/order-service/pkg/adapter"
	"github.com/panggggg/order-service/pkg/entity"
	"github.com/panggggg/order-service/pkg/repository"
	"github.com/panggggg/order-service/pkg/usecase"
	"github.com/wisesight/spider-go-utilities/cache"
	"github.com/wisesight/spider-go-utilities/queue"
)

func main() {
	cfg := config.NewConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmqAdapter, err := queue.NewRabbitMQ(cfg.RabbitMQURI, queue.QueueConfig{
		QueueNames:      []string{},
		ErrorQueueNames: []string{},
		ExchangeName:    "save-order",
		ExchangeType:    "direct",
	})
	if err != nil {
		log.Panic(err)
	}

	redisAdapter := cache.NewRedis("localhost", 6379, "root", 0, ctx)

	orderApi := adapter.NewOrderAPI(cfg)

	orderRepo := repository.NewSaveOrder(orderApi)

	orderUsecase := usecase.NewSaveOrder(orderRepo, redisAdapter)

	jobs, err := rabbitmqAdapter.Consume(cfg.OrderQueueName, 1)
	if err != nil {
		log.Panic(err)
	}
	for job := range jobs {
		var order entity.Order
		err := json.Unmarshal(job.Body, &order)
		if err != nil {
			fmt.Println(err)
			continue
		}

		isExisted, err := orderUsecase.IsExist(ctx, order)
		if isExisted {
			job.Ack(false)
			continue
		}
		if err != nil {
			log.Panic(err)
		}

		err = orderUsecase.SaveOrderStatus(ctx, order)
		if err != nil {
			log.Panic(err)
		}

		err = orderUsecase.Save(order)
		if err != nil {
			log.Panic(err)
		}
		job.Ack(false)
	}

}
