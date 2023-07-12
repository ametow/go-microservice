package order

import (
	"database/sql"
	"time"

	"yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg"
)

type Order struct {
	ID            uint
	Cost          int32
	Weight        float32
	Region        int32
	CompletedTime sql.NullTime `gorm:"index"`
	CreatedAt     time.Time
	DeliveryHours []OrderDeliveryHours `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // has many
	Courier       courier.OrderCourier `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // has one
	GroupID       sql.NullInt32
	GroupOrder    GroupOrder `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type GroupOrder struct {
	ID        uint
	CourierID uint
	Courier   courier.Courier
	Date      time.Time
	Orders    []Order `gorm:"foreignKey:GroupID"`
}

type OrderDeliveryHours struct {
	ID      uint
	OrderID uint
	Starts  pkg.TIME
	Ends    pkg.TIME
}

type OrderService interface {
	FetchSingleOrder(orderID int) (*OrderDto, error)
	FetchOrders(limit, offset int) ([]OrderDto, error)
	CreateNewOrder(in *CreateOrderRequest) ([]OrderDto, error)
	MarkOrdersComplete(in *CompleteOrderRequestDto) ([]OrderDto, error)
	AssignOrdersToCouriers(date time.Time) ([]pkg.OrderAssignResponse, error)
}

type OrderRepository interface {
	GetOrders(limit, offset int) ([]Order, error)
	GetFreeCouriers(date time.Time) ([]courier.Courier, error)
	GetOrderByID(id int) (*Order, error)
	CreateOrder(order CreateOrderDto) (uint, error)
	CompleteOrder(info CompleteOrder) (*Order, error)
	GetCourierAssignments(courierId int, date time.Time) ([]GroupOrder, error)
	GetUnassignedOrders() ([]Order, error)
	CreateOrderGroup(p GroupOrder) error
}
