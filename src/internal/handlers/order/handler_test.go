package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	orderDomain "yandex-team.ru/bstask/internal/order"
	"yandex-team.ru/bstask/internal/pkg/validators"
)

func TestValidateCompleteOrderReq(t *testing.T) {
	input := orderDomain.CompleteOrderRequestDto{}
	err := validateCompleteOrderReq(&input)
	require.Error(t, err)
	require.EqualError(t, err, orderDomain.ErrZeroOrders.Error())

	input = orderDomain.CompleteOrderRequestDto{CompleteInfo: []orderDomain.CompleteOrder{}}
	err = validateCompleteOrderReq(&input)
	require.Error(t, err)
	require.EqualError(t, err, orderDomain.ErrZeroOrders.Error())

	input = orderDomain.CompleteOrderRequestDto{CompleteInfo: []orderDomain.CompleteOrder{
		{CourierId: 1, OrderId: 1, CompleteTime: time.Now().Format(time.RFC3339)},
	}}
	err = validateCompleteOrderReq(&input)
	require.NoError(t, err)

	input = orderDomain.CompleteOrderRequestDto{CompleteInfo: []orderDomain.CompleteOrder{
		{CourierId: 1, OrderId: 1, CompleteTime: ""},
	}}
	err = validateCompleteOrderReq(&input)
	require.EqualError(t, err, validators.ErrInvalidTime.Error())
}

func TestValidateCompleteOrderccess(t *testing.T) {
	in := orderDomain.CompleteOrder{CourierId: 1, OrderId: 1, CompleteTime: time.Now().Format(time.RFC3339)}
	err := validateCompleteOrderDto(&in)
	require.NoError(t, err)
}
func TestValidateCompleteOrderDto(t *testing.T) {
	cases := []struct {
		name      string
		in        orderDomain.CompleteOrder
		expectErr error
	}{
		{
			"bad_time_wrong",
			orderDomain.CompleteOrder{CompleteTime: "sdfd", CourierId: 1, OrderId: 1},
			validators.ErrInvalidTime,
		},
		{
			"bad_time_empty",
			orderDomain.CompleteOrder{CompleteTime: "", CourierId: 1, OrderId: 1},
			validators.ErrInvalidTime,
		},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			err := validateCompleteOrderDto(&tCase.in)
			require.EqualError(t, tCase.expectErr, err.Error())
		})
	}
}

func TestValidateCreateOrderReq(t *testing.T) {
	input := orderDomain.CreateOrderRequest{}
	err := validateCreateOrderReq(&input)
	require.Error(t, err)

	input = orderDomain.CreateOrderRequest{Orders: []orderDomain.CreateOrderDto{}}
	err = validateCreateOrderReq(&input)
	require.Error(t, err)

	input = orderDomain.CreateOrderRequest{Orders: []orderDomain.CreateOrderDto{
		{
			Weight:        1.1,
			Regions:       233,
			DeliveryHours: []string{"11:00-12:00"},
			Cost:          -200,
		},
	}}
	err = validateCreateOrderReq(&input)
	require.EqualError(t, orderDomain.ErrOrderCost, err.Error())

	input = orderDomain.CreateOrderRequest{Orders: []orderDomain.CreateOrderDto{
		{
			Weight:        1.1,
			Regions:       233,
			DeliveryHours: []string{"11:00-12:00"},
			Cost:          200,
		},
	}}
	err = validateCreateOrderReq(&input)
	require.NoError(t, err)
}

func TestValidateCreateOrderccess(t *testing.T) {
	in := orderDomain.CreateOrderDto{
		Weight:        1.2,
		Regions:       23,
		DeliveryHours: []string{"00:11-02:33"},
		Cost:          130,
	}
	err := validateCreateOrderDto(&in)
	require.NoError(t, err)
}

func TestValidateCreateOrderDtoError(t *testing.T) {
	cases := []struct {
		name      string
		in        orderDomain.CreateOrderDto
		expectErr error
	}{
		{
			"cost_zero",
			orderDomain.CreateOrderDto{
				Cost:          0,
				Weight:        1.7,
				Regions:       1000,
				DeliveryHours: []string{"00:04-02:13"},
			},
			orderDomain.ErrOrderCost,
		},
		{
			"cost_negative",
			orderDomain.CreateOrderDto{
				Cost:          -130,
				Weight:        1.7,
				Regions:       1000,
				DeliveryHours: []string{"00:04-02:13"},
			},
			orderDomain.ErrOrderCost,
		},
		{
			"weight_negative",
			orderDomain.CreateOrderDto{
				Cost:          100,
				Weight:        -1.2,
				Regions:       1000,
				DeliveryHours: []string{"00:04-02:13"},
			},
			orderDomain.ErrOrderWeight,
		},
		{
			"weight_zero",
			orderDomain.CreateOrderDto{
				Cost:          100,
				Weight:        0,
				Regions:       1000,
				DeliveryHours: []string{"00:04-02:13"},
			},
			orderDomain.ErrOrderWeight,
		},
		{
			"zero_delivery_hours",
			orderDomain.CreateOrderDto{
				Cost:          100,
				Weight:        1.0,
				Regions:       1000,
				DeliveryHours: []string{},
			},
			validators.ErrInvalidTimeSlice,
		},
		{
			"bad_delivery_hours",
			orderDomain.CreateOrderDto{
				Cost:          100,
				Weight:        1.0,
				Regions:       1000,
				DeliveryHours: []string{"01:0-02:40"},
			},
			validators.ErrInvalidTimeSlice,
		},
		{
			"bad_regions",
			orderDomain.CreateOrderDto{
				Cost:          100,
				Weight:        1.0,
				Regions:       0,
				DeliveryHours: []string{"01:00-02:40"},
			},
			orderDomain.ErrOrderRegions,
		},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			err := validateCreateOrderDto(&tCase.in)
			require.EqualError(t, err, tCase.expectErr.Error())
		})
	}
}
