package courier

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg"
	mock_courier "yandex-team.ru/bstask/internal/pkg/repository/courier/mocks"
)

func TestSingleCourierFetchSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{ID: 5}, nil).Times(1)

	service := NewCourierService(repo)
	_, err := service.FetchSingleCourier(1)
	require.NoError(t, err)
}
func TestSingleCourierFetchFails(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)

	repo.EXPECT().GetCourierByID(1).Return(nil, errors.New("db is down")).Times(1)

	service := NewCourierService(repo)
	_, err := service.FetchSingleCourier(1)
	require.Error(t, err)
}

func TestCreateCourier(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)

	createCour := courier.CreateCourierDto{
		CourierType:  "FOOT",
		Regions:      []int32{4},
		WorkingHours: []string{"14:00-16:00"},
	}

	input := &courier.CreateCourierRequest{
		Couriers: []courier.CreateCourierDto{createCour},
	}
	repo.EXPECT().CreateCourier(createCour).Return(uint(1), nil).Times(1)

	service := NewCourierService(repo)
	_, err := service.CreateNewCouriers(input)
	require.NoError(t, err)
}

func TestFetchCouriers(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)
	repo.EXPECT().GetCouriers(10, 0).Return([]courier.Courier{}, nil).Times(1)
	service := NewCourierService(repo)
	_, err := service.FetchCouriers(10, 0)
	require.NoError(t, err)
}

func TestFetchCourierMetadata(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)
	startDate := time.Now()
	endDate := time.Now()
	courierId := 1

	expected := []courier.OrderCourier{
		{
			OrderID:   1,
			CourierID: uint64(courierId),
			Order: courier.Order{
				Cost: 120,
			},
		},
	}
	repo.EXPECT().GetCourierByID(courierId).Return(&courier.Courier{ID: uint(courierId),
		WorkingHours: []courier.CourierWorkingHours{
			{
				Starts: pkg.TIME{},
				Ends:   pkg.TIME{},
			},
		}}, nil).Times(1)
	repo.EXPECT().GetCourierOrders(courierId, startDate, endDate).Return(expected, nil).Times(1)

	service := NewCourierService(repo)
	_, err := service.FetchCourierMetaData(courierId, startDate, endDate)

	require.NoError(t, err)
}

func TestFetchCouriersAssignments(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)
	date := time.Now()
	courierId := 1

	expected := []courier.GroupOrder{
		{
			ID:        21,
			CourierID: uint(courierId),
			Date:      date,
			Orders: []courier.Order{
				{
					ID:      23,
					Cost:    120,
					Weight:  2.3,
					Region:  4,
					GroupID: 2,
					DeliveryHours: []courier.OrderDeliveryHours{
						{
							Starts: pkg.TIME{},
							Ends:   pkg.TIME{},
						},
					},
				},
			},
		},
	}
	repo.EXPECT().GetCouriersWithOrdersForDate(date, courierId).Return([]courier.Courier{
		{
			ID: uint(courierId),
		},
	}, nil).Times(1)
	repo.EXPECT().GetCourierAssignments(courierId, date).Return(expected, nil).Times(1)

	service := NewCourierService(repo)
	_, err := service.FetchCouriersAssignments(date, courierId)

	require.NoError(t, err)
}
