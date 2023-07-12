package courier

import (
	"database/sql"
	"fmt"
	"time"

	"yandex-team.ru/bstask/internal/pkg"
)

type CourierType uint8

var CourierTypes = map[string]CourierType{
	"FOOT": FOOT,
	"BIKE": BIKE,
	"AUTO": AUTO,
}

const (
	FOOT CourierType = iota + 1
	BIKE
	AUTO
)

type Courier struct {
	ID              uint   `gorm:"primarykey"`
	Type            string `gorm:"size:10; index"`
	CreatedAt       time.Time
	Regions         []CourierRegions      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // has many
	WorkingHours    []CourierWorkingHours `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // has many
	DeliveredOrders []OrderCourier        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // has many
	GroupOrders     []GroupOrder          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // has many
}

type CourierRegions struct {
	ID        uint `gorm:"primarykey"`
	CourierID uint
	Number    int32
}

type CourierWorkingHours struct {
	ID        uint `gorm:"primarykey"`
	CourierID uint
	Starts    pkg.TIME
	Ends      pkg.TIME
}

type OrderCourier struct {
	OrderID       uint64    `gorm:"primaryKey;autoIncrement:false;unique"` // composite primary key
	CourierID     uint64    `gorm:"primaryKey;autoIncrement:false"`        // composite primary key
	CompletedTime time.Time `gorm:"index"`
	Order         Order
}

type GroupOrder struct {
	ID        uint
	CourierID uint
	Date      time.Time
	Orders    []Order `gorm:"foreignKey:GroupID"`
}

type Order struct {
	ID            uint
	Cost          int32
	Weight        float32
	Region        int32
	GroupID       uint
	CompletedTime sql.NullTime
	DeliveryHours []OrderDeliveryHours `gorm:"foreignKey:OrderID"`
}

type OrderDeliveryHours struct {
	OrderID uint
	Starts  pkg.TIME
	Ends    pkg.TIME
}

func (hours OrderDeliveryHours) ToString() string {
	startV, _ := hours.Starts.Value()
	endV, _ := hours.Ends.Value()
	return fmt.Sprintf("%v-%v", startV, endV)
}

type CourierService interface {
	FetchCouriers(limit, offset int) (*GetCouriersResponse, error)
	FetchSingleCourier(id int) (*CourierDto, error)
	CreateNewCouriers(req *CreateCourierRequest) (*CreateCouriersResponse, error)
	FetchCourierMetaData(courierId int, startDate, endDate time.Time) (*GetCourierMetaInfoResponse, error)
	FetchCouriersAssignments(date time.Time, courierId int) (*pkg.OrderAssignResponse, error)
}

type CourierRepository interface {
	GetCouriers(limit, offset int) ([]Courier, error)
	GetCourierByID(id int) (*Courier, error)
	CreateCourier(courier CreateCourierDto) (uint, error)
	GetCourierOrders(courierId int, startDate, endDate time.Time) ([]OrderCourier, error)
	GetCourierAssignments(courierId int, date time.Time) ([]GroupOrder, error)
	GetCouriersWithOrdersForDate(date time.Time, courierId int) ([]Courier, error)
}
