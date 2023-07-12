package pkg

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type InternalErrorResponse struct {
}

func (InternalErrorResponse) Error() string {
	return "internal error occured"
}

type BadRequestResponse struct {
}

func (BadRequestResponse) Error() string {
	return "bad request"
}

type NotFoundResponse struct {
}

func (NotFoundResponse) Error() string {
	return "resource not found"
}

type OrderDto struct {
	Cost          int32    `json:"cost"`
	DeliveryHours []string `json:"delivery_hours"`
	OrderId       int64    `json:"order_id"`
	Regions       int32    `json:"regions"`
	Weight        float32  `json:"weight"`
	CompletedTime string   `json:"completed_time,omitempty"`
}
type GroupOrders struct {
	GroupOrderId int64      `json:"group_order_id"`
	Orders       []OrderDto `json:"orders"`
}

type CouriersGroupOrders struct {
	CourierId int64         `json:"courier_id"`
	Orders    []GroupOrders `json:"orders"`
}

type OrderAssignResponse struct {
	Date     string                `json:"date"`
	Couriers []CouriersGroupOrders `json:"couriers"`
}

// TIME stores only time in db
type TIME time.Time

func (j *TIME) Scan(value interface{}) error {
	bytes, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal TIME value:", value))
	}
	t, err := time.Parse("15:04:05", bytes)
	if err != nil {
		return err
	}
	*j = TIME(t)
	return err
}

func (j TIME) Value() (driver.Value, error) {
	return time.Time(j).Format("15:04"), nil
}

type JSON json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}
