package courier

import (
	"fmt"
	"time"

	"yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg"
)

type courierService struct {
	repo courier.CourierRepository
}

func NewCourierService(r courier.CourierRepository) *courierService {
	return &courierService{r}
}

func (s *courierService) FetchSingleCourier(id int) (*courier.CourierDto, error) {
	c, err := s.repo.GetCourierByID(id)
	if err != nil {
		return nil, err
	}
	response := new(courier.CourierDto)
	return response.FromModel(c), nil
}

func (s *courierService) CreateNewCouriers(req *courier.CreateCourierRequest) (*courier.CreateCouriersResponse, error) {
	response := courier.CreateCouriersResponse{}
	for _, c := range req.Couriers {
		id, err := s.repo.CreateCourier(c)
		if err != nil {
			return nil, err
		}
		response.Couriers = append(response.Couriers, courier.CourierDto{
			CourierId:    int64(id),
			CourierType:  c.CourierType,
			Regions:      c.Regions,
			WorkingHours: c.WorkingHours,
		})
	}
	return &response, nil
}

func (s *courierService) FetchCouriers(limit, offset int) (*courier.GetCouriersResponse, error) {
	couriers, err := s.repo.GetCouriers(limit, offset)
	if err != nil {
		return nil, err
	}

	response := new(courier.GetCouriersResponse)
	response.Couriers = []courier.CourierDto{}
	for _, c := range couriers {
		courierDto := new(courier.CourierDto)
		response.Couriers = append(response.Couriers, *courierDto.FromModel(&c))
	}
	response.Limit = int32(limit)
	response.Offset = int32(offset)
	return response, nil
}

func (s *courierService) FetchCourierMetaData(id int, startDate, endDate time.Time) (*courier.GetCourierMetaInfoResponse, error) {

	c, err := s.repo.GetCourierByID(id)
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, courier.ErrCourierNotFound
	}

	regions := []int32{}
	for _, r := range c.Regions {
		regions = append(regions, r.Number)
	}
	wHours := []string{}
	for _, r := range c.WorkingHours {
		startV, _ := r.Starts.Value()
		endV, _ := r.Ends.Value()
		wHours = append(wHours, fmt.Sprintf("%v-%v", startV, endV))
	}

	ratingCoef := 0
	earningCoef := 0
	switch c.Type {
	case "FOOT":
		earningCoef = 2
		ratingCoef = 3
	case "BIKE":
		earningCoef = 3
		ratingCoef = 2
	case "AUTO":
		earningCoef = 4
		ratingCoef = 1
	}

	response := &courier.GetCourierMetaInfoResponse{}
	response.CourierId = int64(c.ID)
	response.CourierType = c.Type
	response.Regions = regions
	response.WorkingHours = wHours

	courierOrders, err := s.repo.GetCourierOrders(id, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(courierOrders) > 0 { // Задание 2
		var earnings int32
		for _, order := range courierOrders {
			earnings += (int32(earningCoef) * order.Order.Cost)
		}
		var deliveryCount = float64(len(courierOrders))
		var endDateStartDateDiffHours = endDate.Sub(startDate).Hours()
		rating := (deliveryCount / endDateStartDateDiffHours) * float64(ratingCoef)

		response.Earnings = earnings
		response.Rating = int32(rating)
	}
	return response, nil
}

func (s *courierService) FetchCouriersAssignments(date time.Time, courierId int) (*pkg.OrderAssignResponse, error) {
	couriers, err := s.repo.GetCouriersWithOrdersForDate(date, courierId)
	if err != nil {
		return nil, err
	}
	res := new(pkg.OrderAssignResponse)
	res.Date = date.Format("2006-01-02")
	res.Couriers = []pkg.CouriersGroupOrders{}
	for _, c := range couriers {
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
		res.Couriers = append(res.Couriers, pkg.CouriersGroupOrders{
			CourierId: int64(c.ID),
			Orders:    groups,
		})
	}
	return res, nil
}
