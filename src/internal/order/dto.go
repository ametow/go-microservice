package order

import (
	"fmt"
	"time"
)

type CreateOrderDto struct {
	Weight        float32  `json:"weight"`
	Regions       int32    `json:"regions"`
	DeliveryHours []string `json:"delivery_hours"`
	Cost          int32    `json:"cost"`
}

type CreateOrderRequest struct {
	Orders []CreateOrderDto `json:"orders"`
}

type OrderDto struct {
	Cost          int32    `json:"cost"`
	DeliveryHours []string `json:"delivery_hours"`
	OrderId       int64    `json:"order_id"`
	Regions       int32    `json:"regions"`
	Weight        float32  `json:"weight"`
	CompletedTime string   `json:"completed_time,omitempty"`
}

func (c *OrderDto) FromModel(m *Order) *OrderDto {
	dHours := []string{}
	for _, r := range m.DeliveryHours {
		startV, _ := r.Starts.Value()
		endV, _ := r.Ends.Value()
		dHours = append(dHours, fmt.Sprintf("%v-%v", startV, endV))
	}
	o := &OrderDto{
		OrderId:       int64(m.ID),
		Weight:        m.Weight,
		Cost:          m.Cost,
		Regions:       m.Region,
		DeliveryHours: dHours,
	}
	if m.CompletedTime.Valid {
		o.CompletedTime = m.CompletedTime.Time.Format(time.RFC3339)
	}
	return o
}

type CompleteOrder struct {
	CourierId    int64  `json:"courier_id"`
	OrderId      int64  `json:"order_id"`
	CompleteTime string `json:"complete_time"`
}

type CompleteOrderRequestDto struct {
	CompleteInfo []CompleteOrder `json:"complete_info"`
}
