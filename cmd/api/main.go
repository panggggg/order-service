package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/panggggg/order-service/cmd/api/handler"
	"github.com/panggggg/order-service/cmd/api/route"
	"github.com/panggggg/order-service/config"
	"github.com/panggggg/order-service/pkg/repository"
	"github.com/panggggg/order-service/pkg/usecase"
	"github.com/wisesight/spider-go-utilities/database"
	"github.com/wisesight/spider-go-utilities/queue"
)

func main() {
	cfg := config.NewConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongodbClient, err := database.NewMongoDBConnection(ctx, "mongodb://root:root@localhost:27017/?authSource=admin")
	if err != nil {
		panic(err)
	}

	mongodbAdapter := database.NewMongoDB(mongodbClient)

	defer func() {
		if err = mongodbClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	orderCollection := mongodbClient.Database("go_workshop").Collection("order")
	orderStatusCollection := mongodbClient.Database("go_workshop").Collection("order_status")

	queueNames := []string{"order:job"}
	rabbitmqAdapter, err := queue.NewRabbitMQ("amqp://root:root@localhost:5672/", queue.QueueConfig{
		QueueNames:   queueNames,
		ExchangeName: "order",
		ExchangeType: "direct",
	})
	if err != nil {
		log.Panic(err)
	}
	defer rabbitmqAdapter.CleanUp()

	orderRepo := repository.NewOrder(mongodbAdapter, orderCollection, orderStatusCollection, rabbitmqAdapter, cfg)
	orderUsecase := usecase.NewOrder(orderRepo)

	app := echo.New()

	route.NewRoute(cfg, app, handler.NewOrder(orderUsecase))
	err = app.Start(":1234")
	if err != nil {
		go func() {
			if err := app.Start(":1234"); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		defer cancel()
		if err := app.Shutdown(ctx); err != nil {
			app.Logger.Fatal(err)
		}
	}

}
