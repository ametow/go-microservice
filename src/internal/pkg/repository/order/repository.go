package order

import (
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"yandex-team.ru/bstask/internal/courier"
	orderDomain "yandex-team.ru/bstask/internal/order"
	"yandex-team.ru/bstask/internal/pkg"
)

type OrderRepo struct {
	DB *gorm.DB
}

func NewRepo(db *gorm.DB) OrderRepo {
	return OrderRepo{db}
}

func (repo *OrderRepo) GetFreeCouriers(date time.Time) ([]courier.Courier, error) {
	couriers := []courier.Courier{}
	tx := repo.DB.Joins("LEFT JOIN group_order on group_order.courier_id = courier.id and group_order.date = ?", date.Format("2006-01-02")).Preload("Regions").Preload("WorkingHours").Find(&couriers, "group_order.id is null")
	return couriers, tx.Error
}

func (repo *OrderRepo) GetOrders(limit, offset int) ([]orderDomain.Order, error) {
	orders := []orderDomain.Order{}
	tx := repo.DB.Preload("DeliveryHours").Offset(offset).Limit(limit).Find(&orders)
	return orders, tx.Error
}

func (repo *OrderRepo) GetOrderByID(id int) (*orderDomain.Order, error) {
	order := new(orderDomain.Order)
	repo.DB.Preload("DeliveryHours").Find(&order, id)
	if order.ID == 0 {
		return nil, orderDomain.ErrOrderNotFound
	}
	return order, nil
}

func (repo *OrderRepo) CreateOrder(order orderDomain.CreateOrderDto) (uint, error) {
	dHours := []orderDomain.OrderDeliveryHours{}
	for _, v := range order.DeliveryHours {
		hoursStrs := strings.Split(v, "-")
		startTime, _ := time.Parse("15:04", hoursStrs[0])
		endTime, _ := time.Parse("15:04", hoursStrs[1])
		dHours = append(dHours, orderDomain.OrderDeliveryHours{Starts: pkg.TIME(startTime), Ends: pkg.TIME(endTime)})
	}
	orderModel := orderDomain.Order{
		Weight:        order.Weight,
		Cost:          order.Cost,
		Region:        order.Regions,
		DeliveryHours: dHours,
	}
	tx := repo.DB.Save(&orderModel)
	return orderModel.ID, tx.Error
}

func (repo *OrderRepo) CompleteOrder(info orderDomain.CompleteOrder) (*orderDomain.Order, error) {
	tx := repo.DB.Begin()
	cour := courier.Courier{}
	tx.Find(&cour, info.CourierId)
	if cour.ID == 0 {
		tx.Rollback()
		return nil, orderDomain.ErrCourierNotFound
	}
	order := orderDomain.Order{}
	tx.Preload("DeliveryHours").Preload("GroupOrder").Find(&order, info.OrderId)
	if order.ID == 0 {
		tx.Rollback()
		return nil, orderDomain.ErrOrderNotFound
	}
	if !order.GroupID.Valid {
		tx.Rollback()
		return nil, orderDomain.ErrOrderNotAssigned
	}
	if order.GroupOrder.CourierID != uint(info.CourierId) {
		tx.Rollback()
		return nil, orderDomain.ErrOrderAlreadyDelivered
	}
	deliveryOrder := courier.OrderCourier{}
	tx.Find(&deliveryOrder, "order_id = ?", info.OrderId)
	if deliveryOrder.OrderID != 0 && deliveryOrder.CourierID != uint64(info.CourierId) {
		tx.Rollback()
		return nil, orderDomain.ErrOrderAlreadyDelivered
	}

	cTime, err := time.Parse(time.RFC3339, info.CompleteTime)
	if err != nil {
		tx.Rollback()
		return nil, orderDomain.ErrInvalidCompleteTime
	}
	err = order.CompletedTime.Scan(cTime)
	if err != nil {
		log.Println(err)
	}

	orderInfo := courier.OrderCourier{
		CourierID:     uint64(info.CourierId),
		OrderID:       uint64(info.OrderId),
		CompletedTime: cTime,
	}

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Save(&orderInfo).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &order, tx.Commit().Error
}

func (repo *OrderRepo) GetUnassignedOrders() ([]orderDomain.Order, error) {
	orders := []orderDomain.Order{}
	tx := repo.DB.Preload("DeliveryHours").Find(&orders, "completed_time is null and group_id is null")
	return orders, tx.Error
}

func (repo *OrderRepo) GetCourierAssignments(courierId int, date time.Time) ([]orderDomain.GroupOrder, error) {
	grOrders := []orderDomain.GroupOrder{}
	tx := repo.DB.Preload("Orders.DeliveryHours").Find(&grOrders, "courier_id = ? and date = ?", courierId, date)
	return grOrders, tx.Error
}

func (repo *OrderRepo) CreateOrderGroup(p orderDomain.GroupOrder) error {
	tx := repo.DB.Save(&p)
	return tx.Error
}
