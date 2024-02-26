package usecase

import (
	"context"
	"fmt"

	"github.com/panggggg/order-service/pkg/entity"
	"github.com/panggggg/order-service/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order interface {
	GetOrders(ctx context.Context) ([]entity.Order, error)
	GetOrderById(ctx context.Context, orderId string) (entity.Order, error)
	CreateOrder(ctx context.Context, order entity.Order) (*primitive.ObjectID, error)
	UpdateOrder(ctx context.Context, orderId string, updateData entity.Order) (bool, error)
	CreateOrderStatus(ctx context.Context, order entity.OrderStatus) (*primitive.ObjectID, error)
	SendOrdersToQueue(ctx context.Context, order []string) error
}

type order struct {
	orderRepo repository.Order
}

func NewOrder(orderRepo repository.Order) Order {
	return &order{
		orderRepo,
	}
}

func (r order) GetOrders(ctx context.Context) ([]entity.Order, error) {
	return r.orderRepo.GetOrders(ctx)
}

func (r order) GetOrderById(ctx context.Context, orderId string) (entity.Order, error) {
	return r.orderRepo.GetOrderById(ctx, orderId)
}

func (r order) CreateOrder(ctx context.Context, order entity.Order) (*primitive.ObjectID, error) {
	return r.orderRepo.CreateOrder(ctx, order)
}

func (r order) UpdateOrder(ctx context.Context, orderId string, updateData entity.Order) (bool, error) {
	// wait group
	fmt.Println(updateData)
	id, err := r.orderRepo.CreateOrderStatus(ctx, updateData)
	fmt.Println(id)
	if err != nil {
		return false, err
	}
	_, err = r.orderRepo.UpdateOrder(ctx, orderId, updateData)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r order) CreateOrderStatus(ctx context.Context, order entity.OrderStatus) (*primitive.ObjectID, error) {
	// return r.orderRepo.CreateOrderStatus(ctx, order)
	return nil, nil
}

func (r order) SendOrdersToQueue(ctx context.Context, order []string) error {
	return r.orderRepo.SendOrderToQueue(ctx, order)
}
