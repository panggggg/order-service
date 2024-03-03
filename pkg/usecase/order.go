package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/panggggg/order-service/pkg/entity"
	"github.com/panggggg/order-service/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order interface {
	GetOrders(ctx context.Context) ([]entity.Order, error)
	GetOrderById(ctx context.Context, orderId string) (entity.Order, error)
	CreateOrder(ctx context.Context, order entity.Order) (*primitive.ObjectID, error)
	Upsert(ctx context.Context, orderId string, updateData entity.Order) (bool, error)
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

func (r order) Upsert(ctx context.Context, orderId string, updateData entity.Order) (bool, error) {
	// wait group
	var wg sync.WaitGroup
	wg.Add(2)

	var UpdateError error

	go func() {
		fmt.Println("Save status")
		defer wg.Done()
		id, err := r.orderRepo.CreateOrderStatus(ctx, updateData)
		fmt.Println(id)
		if err != nil {
			UpdateError = err
			return
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Save Order")
		_, err := r.orderRepo.Upsert(ctx, orderId, updateData)
		if err != nil {
			UpdateError = err
			return
		}
	}()

	wg.Wait()

	if UpdateError != nil {
		return false, UpdateError
	}
	return true, nil
}

func (r order) SendOrdersToQueue(ctx context.Context, order []string) error {
	return r.orderRepo.SendOrderToQueue(ctx, order)
}
