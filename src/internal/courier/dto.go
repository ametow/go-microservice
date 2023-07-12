package courier

import (
	"fmt"
	"strconv"
	"strings"
)

const MINUTESINADAY = 1440

func atoi(input string) int {
	i, _ := strconv.Atoi(input)
	return i
}

type OrderAssignDto struct {
	Id                   int64
	Cost                 int32
	Weight               float32
	Region               int32
	DeliveryTimes        []string
	deliveryTimeRanges   [][2]int
	deliveryTimeToMinute [MINUTESINADAY]int
}

func (c *OrderAssignDto) createDeliveryTime() {
	c.createRangeArray()
	for _, rangeVal := range c.deliveryTimeRanges {
		for i := rangeVal[0]; i <= rangeVal[1]; i++ {
			c.deliveryTimeToMinute[i] = 1
		}
	}
}

func (c *OrderAssignDto) createRangeArray() {
	for _, rangeVal := range c.DeliveryTimes {
		beginEndArr := strings.Split(rangeVal, "-")
		beginMinute := atoi(strings.Split(beginEndArr[0], ":")[0])
		beginMinute = beginMinute*60 + atoi(strings.Split(beginEndArr[0], ":")[1])
		endMinute := atoi(strings.Split(beginEndArr[1], ":")[0])
		endMinute = endMinute*60 + atoi(strings.Split(beginEndArr[1], ":")[1])
		c.deliveryTimeRanges = append(c.deliveryTimeRanges, [2]int{beginMinute, endMinute})
	}
}

func (c *OrderAssignDto) CheckIsWorkingOnMinute(minute int) bool {
	if minute < 0 {
		return false
	}
	return c.deliveryTimeToMinute[minute] != 0
}

func (o *OrderAssignDto) FromModel(payload Order) *OrderAssignDto {
	dhours := []string{}
	for _, h := range payload.DeliveryHours {
		dhours = append(dhours, h.ToString())
	}
	res := &OrderAssignDto{
		Id:            int64(payload.ID),
		Weight:        payload.Weight,
		DeliveryTimes: dhours,
		Region:        payload.Region,
	}
	res.createDeliveryTime()
	return res
}

type CourierAssignDto struct {
	CourierId             int64
	CourierType           string
	Regions               []int32
	WorkingHours          []string
	WorkingHoursNums      [][2]int
	WorkingHoursInMinutes [MINUTESINADAY]int
	MaxWeight             int
	MaxOrders             int
	MaxRegions            int
	TimeTakenFirst        int
	TimeTakenRest         int
}
type CourierList []Courier

func (e CourierList) Len() int {
	return len(e)
}

func (e CourierList) Less(i, j int) bool {
	return e[i].Type < e[j].Type
}

func (e CourierList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (c *CourierAssignDto) CreateWorkTime() {
	c.ConvertHoursToNumHours()
	for _, rangeVal := range c.WorkingHoursNums {
		for i := rangeVal[0]; i <= rangeVal[1]; i++ {
			c.WorkingHoursInMinutes[i] = 1
		}
	}
}

func (c *CourierAssignDto) ConvertHoursToNumHours() {
	for _, hours := range c.WorkingHours {
		beginEndArr := strings.Split(hours, "-")
		beginMinute := atoi(strings.Split(beginEndArr[0], ":")[0])
		beginMinute = beginMinute*60 + atoi(strings.Split(beginEndArr[0], ":")[1])
		endMinute := atoi(strings.Split(beginEndArr[1], ":")[0])
		endMinute = endMinute*60 + atoi(strings.Split(beginEndArr[1], ":")[1])
		c.WorkingHoursNums = append(c.WorkingHoursNums, [2]int{beginMinute, endMinute})
	}
}

func (c *CourierAssignDto) CheckIsWorkingOnMinute(minute int) bool {
	if minute < 0 {
		return false
	}
	return c.WorkingHoursInMinutes[minute] != 0
}

func (c *CourierAssignDto) CheckConds(order OrderAssignDto) bool {
	hasReg := false
	for _, r := range c.Regions {
		if r == order.Region {
			hasReg = true
			break
		}
	}
	if !hasReg {
		return false
	}
	if float32(c.MaxWeight) < order.Weight {
		return false
	}
	return true

}

func (c *CourierAssignDto) FromModel(m *Courier) *CourierAssignDto {
	regions := []int32{}
	for _, r := range m.Regions {
		regions = append(regions, r.Number)
	}
	wHours := []string{}
	for _, r := range m.WorkingHours {
		startV, _ := r.Starts.Value()
		endV, _ := r.Ends.Value()
		wHours = append(wHours, fmt.Sprintf("%v-%v", startV, endV))
	}
	res := &CourierAssignDto{
		CourierId:    int64(m.ID),
		CourierType:  m.Type,
		Regions:      regions,
		WorkingHours: wHours,
	}
	switch m.Type {
	case "FOOT":
		res.MaxRegions = 1
		res.MaxOrders = 2
		res.MaxWeight = 10
		res.TimeTakenFirst = 25
		res.TimeTakenRest = 10
	case "BIKE":
		res.MaxRegions = 2
		res.MaxOrders = 4
		res.MaxWeight = 20
		res.TimeTakenFirst = 12
		res.TimeTakenRest = 8
	case "AUTO":
		res.MaxRegions = 3
		res.MaxOrders = 7
		res.MaxWeight = 40
		res.TimeTakenFirst = 8
		res.TimeTakenRest = 4
	}
	res.CreateWorkTime()
	return res
}

type CreateCourierDto struct {
	CourierType  string   `json:"courier_type"`
	Regions      []int32  `json:"regions"`
	WorkingHours []string `json:"working_hours"`
}

type CreateCourierRequest struct {
	Couriers []CreateCourierDto `json:"couriers"`
}

type CourierDto struct {
	CourierId    int64    `json:"courier_id"`
	CourierType  string   `json:"courier_type"`
	Regions      []int32  `json:"regions"`
	WorkingHours []string `json:"working_hours"`
}

func (c *CourierDto) FromModel(m *Courier) *CourierDto {
	regions := []int32{}
	for _, r := range m.Regions {
		regions = append(regions, r.Number)
	}
	wHours := []string{}
	for _, r := range m.WorkingHours {
		startV, _ := r.Starts.Value()
		endV, _ := r.Ends.Value()
		wHours = append(wHours, fmt.Sprintf("%v-%v", startV, endV))
	}
	return &CourierDto{
		CourierId:    int64(m.ID),
		CourierType:  m.Type,
		Regions:      regions,
		WorkingHours: wHours,
	}
}

type CreateCouriersResponse struct {
	Couriers []CourierDto `json:"couriers"`
}

type GetCouriersResponse struct {
	Couriers []CourierDto `json:"couriers"`
	Limit    int32        `json:"limit"`
	Offset   int32        `json:"offset"`
}

type GetCourierMetaInfoResponse struct {
	CourierId    int64    `json:"courier_id"`
	CourierType  string   `json:"courier_type"`
	Regions      []int32  `json:"regions"`
	WorkingHours []string `json:"working_hours"`
	Rating       int32    `json:"rating,omitempty"`
	Earnings     int32    `json:"earnings,omitempty"`
}
