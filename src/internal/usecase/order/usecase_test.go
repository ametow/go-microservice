package order

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/order"
	"yandex-team.ru/bstask/internal/pkg"
	mock_order "yandex-team.ru/bstask/internal/pkg/repository/order/mocks"
)

func TestFetchSingleOrder(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := NewOrderService(repo)
	orderId := 1
	repo.EXPECT().GetOrderByID(orderId).Return(&order.Order{ID: 1}, nil)

	_, err := service.FetchSingleOrder(orderId)

	require.NoError(t, err)
}

func TestFetchOrders(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := NewOrderService(repo)
	limit := 10
	offset := 0
	repo.EXPECT().GetOrders(limit, offset).Return([]order.Order{
		{
			ID:     1,
			Cost:   120,
			Weight: 2.3,
			Region: 4,
		},
	}, nil)

	_, err := service.FetchOrders(limit, offset)

	require.NoError(t, err)
}

func TestCreateNewOrder(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := NewOrderService(repo)
	oneDto := order.CreateOrderDto{
		Weight:        4.5,
		Regions:       5,
		DeliveryHours: []string{},
		Cost:          120,
	}
	repo.EXPECT().CreateOrder(oneDto).Return(uint(1), nil)

	_, err := service.CreateNewOrder(&order.CreateOrderRequest{
		Orders: []order.CreateOrderDto{
			oneDto,
		},
	})

	require.NoError(t, err)
}

func TestMakeOrderComplete(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := NewOrderService(repo)
	oneDto := order.CompleteOrder{
		CourierId: 1,
		OrderId:   1,
	}
	repo.EXPECT().CompleteOrder(oneDto).Return(&order.Order{}, nil).Times(1)

	_, err := service.MarkOrdersComplete(&order.CompleteOrderRequestDto{
		CompleteInfo: []order.CompleteOrder{
			oneDto,
		},
	})

	require.NoError(t, err)
}

func TestAssignOrdersToCouriers(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := NewOrderService(repo)
	date := time.Now()
	startsAt, _ := time.Parse("15:04:05", "12:00:00")
	endsAt, _ := time.Parse("15:04:05", "16:00:00")
	courierId := 1
	unassignedOrders := []order.Order{
		{
			ID:            1,
			Cost:          120,
			Weight:        2.3,
			Region:        23,
			DeliveryHours: []order.OrderDeliveryHours{{Starts: pkg.TIME(startsAt), Ends: pkg.TIME(endsAt)}},
		},
	}
	couriersDb := []courier.Courier{
		{
			ID:           uint(courierId),
			Type:         "FOOT",
			Regions:      []courier.CourierRegions{{Number: 23}},
			WorkingHours: []courier.CourierWorkingHours{{Starts: pkg.TIME(startsAt), Ends: pkg.TIME(endsAt)}},
		},
	}
	repo.EXPECT().GetUnassignedOrders().Return(unassignedOrders, nil).Times(1)
	repo.EXPECT().GetFreeCouriers(date).Return(couriersDb, nil).Times(1)
	repo.EXPECT().CreateOrderGroup(order.GroupOrder{
		CourierID: uint(courierId),
		Date:      date,
		Orders:    []order.Order{{ID: 1}},
	}).Return(nil).Times(1)
	repo.EXPECT().GetCourierAssignments(courierId, date).Return([]order.GroupOrder{
		{
			ID:        1,
			CourierID: uint(courierId),
			Date:      date,
			Orders:    unassignedOrders,
		},
	}, nil).Times(1)

	_, err := service.AssignOrdersToCouriers(date)

	require.NoError(t, err)
}
