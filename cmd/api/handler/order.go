package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/panggggg/order-service/pkg/entity"
	"github.com/panggggg/order-service/pkg/usecase"
)

type Order interface {
	GetOrders(c echo.Context) error
	GetOrderById(c echo.Context) error
	CreateOrder(c echo.Context) error
	UpdateOrder(c echo.Context) error
	UploadCsvFile(c echo.Context) error
}

type order struct {
	orderUsecase usecase.Order
}

func NewOrder(orderUsecase usecase.Order) Order {
	return &order{
		orderUsecase: orderUsecase,
	}
}

func (o order) GetOrders(c echo.Context) error {
	if o.orderUsecase == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "orderUsecase is not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := o.orderUsecase.GetOrders(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (o order) GetOrderById(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orderId := c.Param("id")
	res, err := o.orderUsecase.GetOrderById(ctx, orderId)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (o order) CreateOrder(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	order := entity.Order{}
	if err := c.Bind(&order); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	_, err := o.orderUsecase.CreateOrder(ctx, order)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, order)
}

func (o order) UpdateOrder(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orderId := c.Param("id")
	var updateData entity.Order
	if err := c.Bind(&updateData); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	_, err := o.orderUsecase.UpdateOrder(ctx, orderId, updateData)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, updateData)
}

func (o order) UploadCsvFile(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	file, err := c.FormFile("order")
	if err != nil {
		log.Panic(err)
	}

	value, err := file.Open()
	if err != nil {
		log.Panic(err)
	}
	defer value.Close()

	csvData, err := io.ReadAll(value)
	if err != nil {
		log.Panic(err)
	}

	data := bytes.NewReader(csvData)
	r := csv.NewReader(data)
	orders, err := r.ReadAll()
	if err != nil {
		log.Panic(err)
	}
	for _, v := range orders[1:] {
		fmt.Println(v)
		err = o.orderUsecase.SendOrdersToQueue(ctx, v)
		if err != nil {
			log.Panic(err)
		}
	}

	return c.String(http.StatusOK, "File uploaded successfully")
}
