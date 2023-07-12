package courier

import (
	"testing"

	"github.com/stretchr/testify/require"
	courierDomain "yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg/validators"
)

func TestValidateCreateCourierReq(t *testing.T) {
	input := courierDomain.CreateCourierRequest{}
	err := validateCreateCourierReq(&input)
	require.Error(t, err)
	require.EqualError(t, err, courierDomain.ErrZeroLengthCouriers.Error())

	input = courierDomain.CreateCourierRequest{Couriers: []courierDomain.CreateCourierDto{
		{
			CourierType:  "FOOT",
			Regions:      []int32{12, 23},
			WorkingHours: []string{"00:05-01:14"},
		},
	}}
	err = validateCreateCourierReq(&input)
	require.NoError(t, err)

	input = courierDomain.CreateCourierRequest{Couriers: []courierDomain.CreateCourierDto{
		{
			CourierType:  "BAD",
			Regions:      []int32{12, 23},
			WorkingHours: []string{"00:05-01:14"},
		},
	}}
	err = validateCreateCourierReq(&input)
	require.EqualError(t, courierDomain.ErrCourierBadType, err.Error())
}

func TestValidateCreateCourierDto(t *testing.T) {
	cases := []struct {
		name      string
		in        courierDomain.CreateCourierDto
		expectErr error
	}{
		{
			"bad_type",
			courierDomain.CreateCourierDto{CourierType: "SHIP"},
			courierDomain.ErrCourierBadType,
		},
		{
			"empty_regions",
			courierDomain.CreateCourierDto{CourierType: "FOOT"},
			courierDomain.ErrCourierBadRegions,
		},
		{
			"negative regions",
			courierDomain.CreateCourierDto{CourierType: "FOOT", Regions: []int32{-12}},
			courierDomain.ErrCourierBadRegions,
		},
		{
			"empty working hours",
			courierDomain.CreateCourierDto{CourierType: "FOOT", Regions: []int32{12}},
			courierDomain.ErrCourierBadWorkingHours,
		},
		{
			"bad working hours",
			courierDomain.CreateCourierDto{CourierType: "FOOT", Regions: []int32{12}, WorkingHours: []string{"0:00-23:00"}},
			validators.ErrInvalidTimeSlice,
		},
		{
			"invalid working hours",
			courierDomain.CreateCourierDto{CourierType: "FOOT", Regions: []int32{12}, WorkingHours: []string{"00:00-24:00"}},
			validators.ErrInvalidTime,
		},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			err := validateCreateCourierDto(&tCase.in)
			require.EqualError(t, err, tCase.expectErr.Error())
		})
	}
}
