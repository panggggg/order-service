package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/panggggg/order-service/config"
	"github.com/panggggg/order-service/pkg/entity"
	"github.com/wisesight/spider-go-utilities/database"
	"github.com/wisesight/spider-go-utilities/queue"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Order interface {
	GetOrders(ctx context.Context) ([]entity.Order, error)
	GetOrderById(ctx context.Context, orderId string) (entity.Order, error)
	CreateOrder(ctx context.Context, order entity.Order) (*primitive.ObjectID, error)
	UpdateOrder(ctx context.Context, orderId string, updateData entity.Order) (bool, error)
	CreateOrderStatus(ctx context.Context, order entity.Order) (*primitive.ObjectID, error)
	SendOrderToQueue(ctx context.Context, order []string) error
}

type order struct {
	mongodbAdapter        database.MongoDB
	orderCollection       database.MongoCollection
	rabbitmqAdapter       queue.RabbitMQ
	orderStatusCollection database.MongoCollection
	config                config.Config
}

func NewOrder(mongoDBAdapter database.MongoDB, orderCollection database.MongoCollection, orderStatusCollection database.MongoCollection, rabbitmqAdapter queue.RabbitMQ, config config.Config) Order {
	return &order{
		mongodbAdapter:        mongoDBAdapter,
		orderCollection:       orderCollection,
		orderStatusCollection: orderStatusCollection,
		rabbitmqAdapter:       rabbitmqAdapter,
		config:                config,
	}
}

func (o order) GetOrders(ctx context.Context) ([]entity.Order, error) {
	var result []entity.Order
	query := bson.M{
		"status": "pending",
	}
	err := o.mongodbAdapter.Find(ctx, o.orderCollection, &result, query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o order) GetOrderById(ctx context.Context, orderId string) (entity.Order, error) {
	var result entity.Order
	query := bson.M{
		"_id": "order_" + orderId,
	}
	err := o.mongodbAdapter.FindOne(ctx, o.orderCollection, &result, query)
	if err != nil {
		return entity.Order{}, err
	}
	return result, nil
}

func (o order) CreateOrder(ctx context.Context, order entity.Order) (*primitive.ObjectID, error) {
	formattedOrder := entity.Order{
		OrderId: order.OrderId,
		Status:  order.Status,
	}
	result, err := o.mongodbAdapter.InsertOne(ctx, o.orderCollection, formattedOrder)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o order) UpdateOrder(ctx context.Context, orderId string, updateData entity.Order) (bool, error) {
	query := bson.M{
		"_id": "order_" + orderId,
	}
	update := bson.M{
		"$set": updateData,
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
		"$currentDate": bson.M{
			"updated_at": true,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := o.mongodbAdapter.UpdateOne(ctx, o.orderCollection, query, update, opts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (o order) CreateOrderStatus(ctx context.Context, order entity.Order) (*primitive.ObjectID, error) {
	formatData := map[string]interface{}{
		"order_id":   order.OrderId,
		"status":     order.Status,
		"remark":     order.Remark,
		"created_at": time.Now(),
	}
	result, err := o.mongodbAdapter.InsertOne(ctx, o.orderStatusCollection, formatData)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (o order) SendOrderToQueue(ctx context.Context, order []string) error {
	var formatOrder = entity.OrderStatus{
		OrderId: order[0],
		Status:  order[1],
		Remark:  order[2],
	}
	body, err := json.Marshal(formatOrder)
	if err != nil {
		return err
	}
	fmt.Println(formatOrder)

	err = o.rabbitmqAdapter.Publish(ctx, "order:job", body)
	if err != nil {
		fmt.Println("Cannot publish message")
		return err
	}

	fmt.Println("Publish order to queue success")

	return nil
}
