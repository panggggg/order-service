package repository

import (
	"github.com/panggggg/order-service/pkg/adapter"
	"github.com/panggggg/order-service/pkg/entity"
)

type SaveOrder interface {
	SaveOrderWithId(order entity.Order) error
}

type saveOrder struct {
	orderApiAdapter adapter.OrderAPI
}

func NewSaveOrder(orderApi adapter.OrderAPI) *saveOrder {
	return &saveOrder{orderApiAdapter: orderApi}
}

func (s saveOrder) SaveOrderWithId(order entity.Order) error {
	return s.orderApiAdapter.SaveOrder(order)
}
