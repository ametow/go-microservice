package courier

import (
	"strings"
	"time"

	"gorm.io/gorm"
	courierDomain "yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg"
)

type courierRepo struct {
	DB *gorm.DB
}

func NewRepo(db *gorm.DB) *courierRepo {
	return &courierRepo{db}
}

func (repo *courierRepo) GetCourierOrders(courierId int, startDate, endDate time.Time) ([]courierDomain.OrderCourier, error) {
	res := []courierDomain.OrderCourier{}
	tx := repo.DB.Joins("Order", repo.DB.Select("cost")).Find(&res, "order_courier.courier_id = ? and order_courier.completed_time >= ? and order_courier.completed_time < ?", courierId, startDate, endDate)
	return res, tx.Error
}

func (repo *courierRepo) GetCouriersWithOrdersForDate(date time.Time, courierId int) ([]courierDomain.Courier, error) {
	couriers := []courierDomain.Courier{}
	query := repo.DB.Select("courier.id").Joins("JOIN group_order on group_order.courier_id = courier.id and group_order.date = ?", date.Format("2006-01-02")).Group("courier.id").Session(&gorm.Session{})
	if courierId > 0 {
		query = query.Where("courier.id = ?", courierId)
	}
	tx := query.Find(&couriers)
	return couriers, tx.Error
}

func (repo *courierRepo) GetCouriers(limit, offset int) ([]courierDomain.Courier, error) {
	couriers := []courierDomain.Courier{}
	tx := repo.DB.Preload("Regions").Preload("WorkingHours").Offset(offset).Limit(limit).Find(&couriers)
	return couriers, tx.Error
}

func (repo *courierRepo) GetCourierByID(id int) (*courierDomain.Courier, error) {
	courier := new(courierDomain.Courier)
	tx := repo.DB.Preload("Regions").Preload("WorkingHours").Find(&courier, id)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if courier.ID == 0 {
		return nil, courierDomain.ErrCourierNotFound
	}
	return courier, nil
}

func (repo *courierRepo) CreateCourier(courier courierDomain.CreateCourierDto) (uint, error) {
	wHours := []courierDomain.CourierWorkingHours{}
	for _, v := range courier.WorkingHours {
		hoursStrs := strings.Split(v, "-")
		startTime, _ := time.Parse("15:04", hoursStrs[0])
		endTime, _ := time.Parse("15:04", hoursStrs[1])
		wHours = append(wHours, courierDomain.CourierWorkingHours{Starts: pkg.TIME(startTime), Ends: pkg.TIME(endTime)})
	}
	regions := []courierDomain.CourierRegions{}
	for _, r := range courier.Regions {
		regions = append(regions, courierDomain.CourierRegions{Number: r})
	}
	c := courierDomain.Courier{
		Type:         courier.CourierType,
		WorkingHours: wHours,
		Regions:      regions,
	}
	tx := repo.DB.Save(&c)
	return c.ID, tx.Error
}

func (repo *courierRepo) GetCourierAssignments(courierId int, date time.Time) ([]courierDomain.GroupOrder, error) {
	grOrders := []courierDomain.GroupOrder{}
	tx := repo.DB.Preload("Orders.DeliveryHours").Find(&grOrders, "courier_id = ? and date = ?", courierId, date)
	return grOrders, tx.Error
}
