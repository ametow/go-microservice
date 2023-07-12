package order

import (
	"fmt"
	"sort"
	"time"

	"yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/order"
	"yandex-team.ru/bstask/internal/pkg"
)

type orderService struct {
	repo order.OrderRepository
}

func NewOrderService(r order.OrderRepository) *orderService {
	return &orderService{r}
}

func (s *orderService) FetchSingleOrder(orderID int) (*order.OrderDto, error) {
	o, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}
	if o.ID == 0 {
		return nil, order.ErrOrderNotFound
	}
	response := new(order.OrderDto)
	return response.FromModel(o), nil
}

func (s *orderService) FetchOrders(limit, offset int) ([]order.OrderDto, error) {
	orders, err := s.repo.GetOrders(limit, offset)
	if err != nil {
		return nil, err
	}

	response := []order.OrderDto{}
	for _, o := range orders {
		orderDto := new(order.OrderDto)
		response = append(response, *orderDto.FromModel(&o))
	}
	return response, nil
}

func (s *orderService) CreateNewOrder(in *order.CreateOrderRequest) ([]order.OrderDto, error) {
	response := []order.OrderDto{}
	for _, o := range in.Orders {
		id, err := s.repo.CreateOrder(o)
		if err != nil {
			return nil, err
		}
		response = append(response, order.OrderDto{
			Cost:          o.Cost,
			DeliveryHours: o.DeliveryHours,
			OrderId:       int64(id),
			Regions:       o.Regions,
			Weight:        o.Weight,
		})
	}
	return response, nil
}

func (s *orderService) MarkOrdersComplete(in *order.CompleteOrderRequestDto) ([]order.OrderDto, error) {
	response := []order.OrderDto{}
	orders := []order.Order{}
	for _, cInfo := range in.CompleteInfo {
		order, err := s.repo.CompleteOrder(cInfo)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}
	for _, o := range orders {
		orderDto := order.OrderDto{}
		response = append(response, *orderDto.FromModel(&o))
	}
	return response, nil
}

// Задание 4
func (s *orderService) AssignOrdersToCouriers(date time.Time) ([]pkg.OrderAssignResponse, error) {
	unassignOrdersDb, err := s.repo.GetUnassignedOrders()
	if err != nil {
		return nil, err
	}
	couriersDb, err := s.repo.GetFreeCouriers(date)
	if err != nil {
		return nil, err
	}

	sort.Sort(courier.CourierList(couriersDb))

	var couriers []courier.CourierAssignDto
	for _, c := range couriersDb {
		p := courier.CourierAssignDto{}
		couriers = append(couriers, *p.FromModel(&c))
	}

	var orders []courier.OrderAssignDto
	for _, o := range unassignOrdersDb {
		p := courier.OrderAssignDto{}
		ordHours := []courier.OrderDeliveryHours{}
		for _, h := range o.DeliveryHours {
			ordHours = append(ordHours, courier.OrderDeliveryHours{
				Starts: h.Starts,
				Ends:   h.Ends,
			})
		}
		ord := courier.Order{
			ID:            o.ID,
			Cost:          o.Cost,
			Weight:        o.Weight,
			Region:        o.Region,
			DeliveryHours: ordHours,
		}
		orders = append(orders, *p.FromModel(ord))
	}

	type OrderGroup struct {
		deliveryTimeRange []int
		orders            []int
		label             int
		courierType       string
		weight            float64
	}

	type courierType struct {
		timeTakenForFirst int
		timeTakenForRest  int
		maxWeight         int
		maxOrders         int
	}

	TYPEMAP := map[string]courierType{
		"FOOT": {
			timeTakenForFirst: 25,
			timeTakenForRest:  10,
			maxWeight:         10,
			maxOrders:         2,
		},
		"BIKE": {
			timeTakenForFirst: 12,
			timeTakenForRest:  8,
			maxWeight:         20,
			maxOrders:         4,
		},
		"AUTO": {
			timeTakenForFirst: 8,
			timeTakenForRest:  4,
			maxWeight:         40,
			maxOrders:         7,
		},
	}

	var (
		courierOrderMatrix = make([][]int, len(couriers))
		minuteCheckers     = make([][]int, courier.MINUTESINADAY)
		maxOrderCount      int
		finalList          []int
		takenOrders        []int
		globalQueue        []OrderGroup
		orderGroups        []OrderGroup
	)

	var courierAcceptedMinutes = func(orderIdx int, courierIdx int) []int {
		result := []int{}
		for minute := 1; minute < courier.MINUTESINADAY; minute++ {
			if orders[orderIdx].CheckIsWorkingOnMinute(minute) &&
				couriers[courierIdx].CheckIsWorkingOnMinute(minute-couriers[courierIdx].TimeTakenFirst) {
				result = append(result, minute)
			}
		}
		return result
	}
	var canGroupTakeOrder = func(groupMinutes []int, orderIdx int, needMinute int) []int {
		result := []int{}
		for index := 0; index < len(groupMinutes); index++ {
			if orders[orderIdx].CheckIsWorkingOnMinute(groupMinutes[index] + needMinute) {
				result = append(result, groupMinutes[index]+needMinute)
			}
		}
		return result
	}

	var canTake = func(groupIndex int, label int) bool {
		minuteCheckerTemp := make([]int, courier.MINUTESINADAY)
		group := orderGroups[groupIndex]
		canTake := false
		needMinute := TYPEMAP[group.courierType].timeTakenForFirst + (TYPEMAP[group.courierType].timeTakenForRest * (group.label - 1))
		for _, minute := range group.deliveryTimeRange {
			if label == 1 {
				for minutes := minute - needMinute; minutes <= minute; minutes++ {
					minuteCheckerTemp[minutes] = 1
				}
				canTake = true
				break
			} else {
				if minuteCheckers[label-1][minute] == 0 && minuteCheckers[label-1][minute-needMinute] == 0 {
					for minutes := minute - needMinute; minutes <= minute; minutes++ {
						minuteCheckerTemp[minutes] = 1
					}
					canTake = true
					break
				}
			}
		}
		orderGroups[groupIndex].deliveryTimeRange = []int{}

		if canTake {
			minuteCheckers[label] = minuteCheckerTemp
		}
		return canTake
	}

	var selectOrders func(groups []int, orders []int, label int)

	selectOrders = func(groups []int, orders []int, label int) {
		for groupIndex := 0; groupIndex < len(orderGroups); groupIndex++ {
			if !contains(groups, groupIndex) && notContainsSomeOrders(orders, orderGroups[groupIndex].orders) && canTake(groupIndex, label+1) {
				selectOrders(append(groups, groupIndex), append(orders, orderGroups[groupIndex].orders...), label+1)
			}
		}
		if maxOrderCount < len(orders) {
			finalList = groups
			maxOrderCount = len(orders)
			takenOrders = orders
		}
	}

	var orderAllGroups = func() {
		maxOrderCount = 0
		finalList = []int{}
		takenOrders = []int{}
		minuteCheckers = make([][]int, courier.MINUTESINADAY)
		for index := 0; index < len(orderGroups); index++ {
			if canTake(index, 1) {
				selectOrders([]int{index}, orderGroups[index].orders, 1)
			}
		}
	}

	var findAllGroups = func(courierIdx int) {
		for len(globalQueue) > 0 {
			group := globalQueue[0]
			globalQueue = globalQueue[1:]
			if len(group.orders) == TYPEMAP[group.courierType].maxOrders {
				orderGroups = append(orderGroups, group)
				continue
			}
			tookOne := false
			for orderIndex := 0; orderIndex < len(orders); orderIndex++ {
				if courierOrderMatrix[courierIdx][orderIndex] == 1 && !contains(group.orders, orderIndex) && group.weight+float64(orders[orderIndex].Weight) <= float64(TYPEMAP[group.courierType].maxWeight) {
					needMinutesForOrder := canGroupTakeOrder(group.deliveryTimeRange, orderIndex, couriers[courierIdx].TimeTakenRest)
					if len(needMinutesForOrder) > 0 {
						tookOne = true
						newOrders := make([]int, len(group.orders)+1)
						copy(newOrders, group.orders)
						newOrders[len(group.orders)] = orderIndex
						globalQueue = append(globalQueue, OrderGroup{deliveryTimeRange: needMinutesForOrder, orders: newOrders, label: group.label + 1, courierType: couriers[courierIdx].CourierType, weight: group.weight + float64(orders[orderIndex].Weight)})
					}
				}
			}
			if !tookOne {
				orderGroups = append(orderGroups, group)
			}
		}
		orderAllGroups()
	}

	var getOrderGroups = func(courierIdx int) {
		globalQueue = []OrderGroup{}
		orderGroups = []OrderGroup{}
		for orderIndex := 0; orderIndex < len(orders); orderIndex++ {
			if courierOrderMatrix[courierIdx][orderIndex] == 1 {
				acceptedMinutes := courierAcceptedMinutes(orderIndex, courierIdx)
				if len(acceptedMinutes) > 0 {
					globalQueue = append(globalQueue, OrderGroup{acceptedMinutes, []int{orderIndex}, 1, couriers[courierIdx].CourierType, float64(orders[orderIndex].Weight)})
				}
			}
		}
		findAllGroups(courierIdx)
	}

	// init matrix
	for i := range courierOrderMatrix {
		courierOrderMatrix[i] = make([]int, len(orders))
	}

	for i := 0; i < len(couriers); i++ {
		for j := 0; j < len(orders); j++ {
			if couriers[i].CheckConds(orders[j]) {
				courierOrderMatrix[i][j] = 1
			}
		}
	}

	assignedCouriers := []courier.Courier{}

	for courierIdx := 0; courierIdx < len(couriers); courierIdx++ {
		getOrderGroups(courierIdx)
		for index := courierIdx + 1; index < len(couriers); index++ {
			for _, orderIdx := range takenOrders {
				courierOrderMatrix[index][orderIdx] = 2
			}
		}
		courierId := couriers[courierIdx].CourierId

		for _, groupIdx := range finalList {
			ordersToAttach := []order.Order{}
			for _, o := range orderGroups[groupIdx].orders {
				ordersToAttach = append(ordersToAttach, order.Order{ID: uint(orders[o].Id)})
			}
			err := s.repo.CreateOrderGroup(order.GroupOrder{
				CourierID: uint(courierId),
				Date:      date,
				Orders:    ordersToAttach,
			})
			if err != nil {
				return nil, err
			}
		}
		if len(finalList) > 0 {
			assignedCouriers = append(assignedCouriers, courier.Courier{ID: uint(courierId)})
		}
	}

	response := []pkg.OrderAssignResponse{}
	assignResponse := pkg.OrderAssignResponse{}
	assignResponse.Date = date.Format("2006-01-02")
	assignResponse.Couriers = []pkg.CouriersGroupOrders{}
	for _, c := range assignedCouriers {
		groups := []pkg.GroupOrders{}
		groupOrders, _ := s.repo.GetCourierAssignments(int(c.ID), date)
		for _, group := range groupOrders {
			orderDtos := []pkg.OrderDto{}
			for _, order := range group.Orders {
				dHours := []string{}
				for _, r := range order.DeliveryHours {
					startV, _ := r.Starts.Value()
					endV, _ := r.Ends.Value()
					dHours = append(dHours, fmt.Sprintf("%v-%v", startV, endV))
				}
				orderDto := pkg.OrderDto{
					Cost:          order.Cost,
					Weight:        order.Weight,
					OrderId:       int64(order.ID),
					DeliveryHours: dHours,
					Regions:       order.Region,
				}
				if order.CompletedTime.Valid {
					orderDto.CompletedTime = order.CompletedTime.Time.Format(time.RFC3339)
				}
				orderDtos = append(orderDtos, orderDto)
			}
			groups = append(groups, pkg.GroupOrders{
				GroupOrderId: int64(group.ID),
				Orders:       orderDtos,
			})
		}
		assignResponse.Couriers = append(assignResponse.Couriers, pkg.CouriersGroupOrders{
			CourierId: int64(c.ID),
			Orders:    groups,
		})
	}
	response = append(response, assignResponse)
	return response, nil
}

func notContainsSomeOrders(orders []int, mustTakeOrders []int) bool {
	for _, id := range mustTakeOrders {
		if contains(orders, id) {
			return false
		}
	}
	return true
}

func contains(arr []int, val int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
