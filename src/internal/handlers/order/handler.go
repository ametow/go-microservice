package order

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	orderDomain "yandex-team.ru/bstask/internal/order"
	"yandex-team.ru/bstask/internal/pkg"
	"yandex-team.ru/bstask/internal/pkg/validators"
)

type OrderHandler struct {
	service orderDomain.OrderService
}

func NewHandler(s orderDomain.OrderService) *OrderHandler {
	h := &OrderHandler{s}
	return h
}

func (h *OrderHandler) Init(e *echo.Echo) {
	g := e.Group("/orders")
	g.GET("", h.getOrders)
	g.GET("/:order_id", h.getOrder)
	g.POST("", h.createOrder)
	g.POST("/assign", h.ordersAssign)
	g.POST("/complete", h.completeOrder)
}

// e.GET("/orders/:order_id", getOrder)
func (h *OrderHandler) getOrder(ctx echo.Context) error {
	orderIdStr := ctx.Param("order_id")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	response, err := h.service.FetchSingleOrder(orderId)
	if err != nil {
		if errors.Is(err, orderDomain.ErrOrderNotFound) {
			return ctx.JSON(http.StatusNotFound, pkg.NotFoundResponse{})
		}
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}
	return ctx.JSON(http.StatusOK, response)
}

// e.GET("/orders", getOrders)
func (h *OrderHandler) getOrders(ctx echo.Context) error {
	limit := ctx.QueryParam("limit")
	if limit == "" {
		limit = "1"
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	offset := ctx.QueryParam("offset")
	if offset == "" {
		offset = "0"
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	response, err := h.service.FetchOrders(limitInt, offsetInt)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}
	return ctx.JSON(http.StatusOK, response)
}

// e.POST("/orders", createOrder)
func (h *OrderHandler) createOrder(ctx echo.Context) error {
	in := new(orderDomain.CreateOrderRequest)
	err := ctx.Bind(in)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	err = validateCreateOrderReq(in)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	response, err := h.service.CreateNewOrder(in)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}
	return ctx.JSON(http.StatusOK, response)
}

// e.POST("/orders/complete", completeOrder)
func (h *OrderHandler) completeOrder(ctx echo.Context) error {
	in := new(orderDomain.CompleteOrderRequestDto)
	err := ctx.Bind(in)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	err = validateCompleteOrderReq(in)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}

	response, err := h.service.MarkOrdersComplete(in)
	if err != nil {
		if errors.Is(err, orderDomain.ErrCourierNotFound) ||
			errors.Is(err, orderDomain.ErrInvalidCompleteTime) ||
			errors.Is(err, orderDomain.ErrOrderNotAssigned) ||
			errors.Is(err, orderDomain.ErrOrderNotFound) ||
			errors.Is(err, orderDomain.ErrOrderAlreadyDelivered) {
			return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
		}
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}

	return ctx.JSON(http.StatusOK, response)
}

// e.POST("/orders/assign", ordersAssign)
func (h *OrderHandler) ordersAssign(ctx echo.Context) error {
	dateFormat := "2006-01-02"
	dateStr := ctx.QueryParam("date")
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		date, _ = time.Parse(dateFormat, time.Now().Format(dateFormat))
	}
	response, err := h.service.AssignOrdersToCouriers(date)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}
	return ctx.JSON(http.StatusCreated, response)
}

func validateCompleteOrderReq(r *orderDomain.CompleteOrderRequestDto) error {
	if len(r.CompleteInfo) == 0 {
		return orderDomain.ErrZeroOrders
	}
	for _, c := range r.CompleteInfo {
		err := validateCompleteOrderDto(&c)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCompleteOrderDto(r *orderDomain.CompleteOrder) error {
	if r.CompleteTime == "" {
		return validators.ErrInvalidTime
	}
	_, err := time.Parse(time.RFC3339, r.CompleteTime)
	if err != nil {
		return validators.ErrInvalidTime
	}
	return nil
}

func validateCreateOrderReq(r *orderDomain.CreateOrderRequest) error {
	if len(r.Orders) == 0 {
		return orderDomain.ErrZeroOrders
	}
	for _, c := range r.Orders {
		err := validateCreateOrderDto(&c)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCreateOrderDto(r *orderDomain.CreateOrderDto) error {
	if r.Cost <= 0 {
		return orderDomain.ErrOrderCost
	}
	if r.Weight <= 0.0 {
		return orderDomain.ErrOrderWeight
	}
	if r.Regions <= 0 {
		return orderDomain.ErrOrderRegions
	}
	if len(r.DeliveryHours) == 0 {
		return validators.ErrInvalidTimeSlice
	}
	err := validators.ValidateHoursSlice(r.DeliveryHours)
	return err
}
