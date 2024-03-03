package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/panggggg/order-service/pkg/entity"
	"github.com/panggggg/order-service/pkg/repository"
	"github.com/wisesight/spider-go-utilities/cache"
)

type SaveOrder interface {
	IsExist(ctx context.Context, order entity.Order) (bool, error)
	SaveOrderStatus(ctx context.Context, order entity.Order) error
	Save(order entity.Order) error
}

type saveOrder struct {
	orderRepo    repository.SaveOrder
	redisAdapter cache.Redis
}

func NewSaveOrder(orderRepo repository.SaveOrder, redisAdapter cache.Redis) *saveOrder {
	return &saveOrder{
		orderRepo:    orderRepo,
		redisAdapter: redisAdapter,
	}
}

func (s saveOrder) IsExist(ctx context.Context, order entity.Order) (bool, error) {
	fmt.Println("Get Redis")
	orderId := order.OrderId
	value, err := s.redisAdapter.Get(orderId)
	if err != nil {
		return false, err
	}
	var orderStatus entity.Order
	err = json.Unmarshal([]byte(*value), &orderStatus)

	if err != nil {
		return false, err
	}

	if value != nil && orderStatus.Status == order.Status {
		return true, nil
	}
	return false, nil
}

func (s saveOrder) SaveOrderStatus(ctx context.Context, order entity.Order) error {
	orderId := order.OrderId
	value, err := json.Marshal(order)
	if err != nil {
		return err
	}
	err = s.redisAdapter.Set(orderId, string(value), 15*24*time.Hour)
	if err != nil {
		return err
	}
	return nil
}

func (s saveOrder) Save(order entity.Order) error {
	return s.orderRepo.SaveOrderWithId(order)
}
