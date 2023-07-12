package courier

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg"
	mock_courier "yandex-team.ru/bstask/internal/pkg/repository/courier/mocks"
	courierService "yandex-team.ru/bstask/internal/usecase/courier"
)

func TestInit(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	h := NewHandler(service)
	h.Init(e)
}
func TestNewHandler(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	NewHandler(service)
}
func TestGetCourierByIdSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{ID: 5}, nil).Times(1)

	courierHandler := CourierHandler{service}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/couriers", nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, courierHandler.getCourierById(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCourierByIdFail(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{}, courier.ErrCourierNotFound).Times(1)

	courierHandler := CourierHandler{service}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/couriers", nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("o")

	require.NoError(t, courierHandler.getCourierById(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/couriers", nil)

	c = e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, courierHandler.getCourierById(c))
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCouriersSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)

	repo.EXPECT().GetCouriers(10, 0).Return([]courier.Courier{}, nil).Times(1)

	courierHandler := CourierHandler{service}

	rec := httptest.NewRecorder()

	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "0")
	req := httptest.NewRequest(http.MethodGet, "/couriers?"+q.Encode(), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, courierHandler.getCouriers(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCouriersInvalidParamsFail(t *testing.T) {

	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	repo.EXPECT().GetCouriers(1, 0).Return([]courier.Courier{
		{
			ID:           1,
			Type:         "FOOT",
			Regions:      []courier.CourierRegions{{ID: 1, CourierID: 1, Number: 1}},
			WorkingHours: []courier.CourierWorkingHours{{ID: 1, Starts: pkg.TIME{}, Ends: pkg.TIME{}}},
		},
	}, nil).Times(1)

	courierHandler := CourierHandler{service}

	tcases := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "invalid_limit",
			input:  `limit=1o&offset=1`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_offset",
			input:  `limit=10&offset=a`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "empty_offset_limit",
			input:  `limit=&offset=`,
			expect: http.StatusOK,
		},
	}
	for _, tCase := range tcases {
		t.Run(tCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/couriers?"+tCase.input, nil)
			c := e.NewContext(req, rec)
			require.NoError(t, courierHandler.getCouriers(c))
			require.Equal(t, tCase.expect, rec.Code)
		})
	}
}

func TestGetCouriersDbDown(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()
	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	repo.EXPECT().GetCouriers(10, 0).Return(nil, errors.New("db is down")).Times(1)

	orderHandler := CourierHandler{service}
	rec := httptest.NewRecorder()
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "0")
	req := httptest.NewRequest(http.MethodGet, "/couriers?"+q.Encode(), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.getCouriers(c))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

var (
	createCourierJson     = `{ "couriers": [ { "courier_type": "BIKE", "regions": [ 12, 23 ], "working_hours": [ "15:00-18:00", "13:23-22:00" ] } ] }`
	expectedResFromCreate = "{\"couriers\":[{\"courier_id\":1,\"courier_type\":\"BIKE\",\"regions\":[12,23],\"working_hours\":[\"15:00-18:00\",\"13:23-22:00\"]}]}\n"
)

func TestCreateCourier(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)

	input := courier.CreateCourierDto{
		CourierType:  "BIKE",
		Regions:      []int32{12, 23},
		WorkingHours: []string{"15:00-18:00", "13:23-22:00"},
	}
	repo.EXPECT().CreateCourier(input).Return(uint(1), nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/couriers", strings.NewReader(createCourierJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	c := e.NewContext(req, rec)
	orderHandler := CourierHandler{service}

	require.NoError(t, orderHandler.createCourier(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, expectedResFromCreate, rec.Body.String())
}

func TestCreateCourierFail(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)

	orderHandler := CourierHandler{service}
	input := courier.CreateCourierDto{
		CourierType:  "FOOT",
		Regions:      []int32{12},
		WorkingHours: []string{"13:00-15:00", "13:23-22:00"},
	}
	repo.EXPECT().CreateCourier(input).Return(uint(0), errors.New("db is down")).Times(1)

	tcases := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "invalid_input",
			input:  `{"couriers":["courier_type":"FOOT","regions":[12],"working_hours":["13:00-15:00","13:23-22:00"]}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_time",
			input:  `{"couriers":[{"courier_type":"FOOT","regions":[12],"working_hours":["3:00-15:00","13:23-22:00"]}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_type",
			input:  `{"couriers":[{"courier_type":"feet","regions":[12],"working_hours":["13:00-15:00","13:23-22:00"]}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "db is down case",
			input:  `{"couriers":[{"courier_type":"FOOT","regions":[12],"working_hours":["13:00-15:00","13:23-22:00"]}]}`,
			expect: http.StatusInternalServerError,
		},
	}
	for _, tCase := range tcases {
		t.Run(tCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/couriers", strings.NewReader(tCase.input))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			require.NoError(t, orderHandler.createCourier(c))
			require.Equal(t, tCase.expect, rec.Code)
		})
	}
}

func TestGetCourierMetaInfoAUTOSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-02"
	endD := "2023-01-04"

	startDate, _ := time.Parse("2006-01-02", startD)
	endDate, _ := time.Parse("2006-01-02", endD)
	nowDate := time.Now()

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{
		ID:           1,
		Type:         "AUTO",
		Regions:      []courier.CourierRegions{{Number: 2, ID: 1}},
		WorkingHours: []courier.CourierWorkingHours{{ID: 3, Starts: pkg.TIME{}, Ends: pkg.TIME{}}},
	}, nil).Times(1)
	repo.EXPECT().GetCourierOrders(1, startDate, endDate).Return([]courier.OrderCourier{
		{
			OrderID:       1,
			CourierID:     1,
			CompletedTime: nowDate,
			Order:         courier.Order{Cost: 120},
		},
	}, nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusOK, rec.Code)
}
func TestGetCourierMetaInfoBIKESuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-02"
	endD := "2023-01-04"

	startDate, _ := time.Parse("2006-01-02", startD)
	endDate, _ := time.Parse("2006-01-02", endD)
	nowDate := time.Now()

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{
		ID:           1,
		Type:         "BIKE",
		Regions:      []courier.CourierRegions{{Number: 2, ID: 1}},
		WorkingHours: []courier.CourierWorkingHours{{ID: 3, Starts: pkg.TIME{}, Ends: pkg.TIME{}}},
	}, nil).Times(1)
	repo.EXPECT().GetCourierOrders(1, startDate, endDate).Return([]courier.OrderCourier{
		{
			OrderID:       1,
			CourierID:     1,
			CompletedTime: nowDate,
			Order:         courier.Order{Cost: 120},
		},
	}, nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusOK, rec.Code)
}
func TestGetCourierMetaInfoFOOTSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-02"
	endD := "2023-01-04"

	startDate, _ := time.Parse("2006-01-02", startD)
	endDate, _ := time.Parse("2006-01-02", endD)
	nowDate := time.Now()

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{
		ID:           1,
		Type:         "FOOT",
		Regions:      []courier.CourierRegions{{Number: 2, ID: 1}},
		WorkingHours: []courier.CourierWorkingHours{{ID: 3, Starts: pkg.TIME{}, Ends: pkg.TIME{}}},
	}, nil).Times(1)
	repo.EXPECT().GetCourierOrders(1, startDate, endDate).Return([]courier.OrderCourier{
		{
			OrderID:       1,
			CourierID:     1,
			CompletedTime: nowDate,
			Order:         courier.Order{Cost: 120},
		},
	}, nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCourierMetaInfoInvalidParamFails(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-02"
	endD := "2023-01-04"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("invalid int")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCourierMetaInfoInvalidQueryParamsFail(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-0a"
	endD := "2023-01-04"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")
	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	startD = "2023-01-02"
	endD = "2023-01-0o"
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)
	c = e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCourierMetaInfoDbFails(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-02"
	endD := "2023-01-04"

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{}, errors.New("db is down")).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

}

func TestGetCourierMetaInfoDb2Fails(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	orderHandler := CourierHandler{service}
	startD := "2023-01-02"
	endD := "2023-01-04"

	startDate, _ := time.Parse("2006-01-02", startD)
	endDate, _ := time.Parse("2006-01-02", endD)

	repo.EXPECT().GetCourierByID(1).Return(&courier.Courier{ID: 5}, nil).Times(1)
	repo.EXPECT().GetCourierOrders(1, startDate, endDate).Return([]courier.OrderCourier{}, errors.New("db is down")).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/meta-info/?startDate=%s&endDate=%s", startD, endD), nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("courier_id")
	c.SetParamValues("1")

	require.NoError(t, orderHandler.courierMetaInfo(c))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

}

func TestCourierAssignmentsSuccess(t *testing.T) {

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	date, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	expectedCouriersWithOrders := []courier.Courier{
		{
			ID:      1,
			Type:    "FOOT",
			Regions: []courier.CourierRegions{},
			WorkingHours: []courier.CourierWorkingHours{
				{
					ID:     1,
					Starts: pkg.TIME{},
					Ends:   pkg.TIME{},
				},
			},
			GroupOrders: []courier.GroupOrder{
				{
					ID:        22,
					CourierID: 1,
					Date:      date,
					Orders: []courier.Order{
						{
							ID:            32,
							Cost:          120,
							Weight:        2.3,
							Region:        23,
							GroupID:       1,
							CompletedTime: sql.NullTime{},
							DeliveryHours: []courier.OrderDeliveryHours{
								{
									Starts: pkg.TIME{},
									Ends:   pkg.TIME{},
								},
							},
						},
					},
				},
			},
		},
	}
	repo.EXPECT().GetCouriersWithOrdersForDate(date, 1).Return(expectedCouriersWithOrders, nil).Times(1)
	cTime := sql.NullTime{}
	err := cTime.Scan(time.Now())
	if err != nil {
		log.Println(err)
	}
	repo.EXPECT().GetCourierAssignments(1, date).Return([]courier.GroupOrder{
		{
			ID:        1,
			CourierID: 1,
			Date:      date,
			Orders: []courier.Order{
				{
					ID:            2,
					Cost:          130,
					Weight:        4.3,
					Region:        32,
					CompletedTime: cTime,
					DeliveryHours: []courier.OrderDeliveryHours{
						{
							Starts: pkg.TIME{},
							Ends:   pkg.TIME{},
						},
					},
				},
			},
		},
	}, nil).Times(1)

	orderHandler := CourierHandler{service}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/assignments?courier_id=%d", 1), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.couriersAssignments(c))
	require.Equal(t, http.StatusOK, rec.Code)
}
func TestCourierAssignmentsInvalidParamsFail(t *testing.T) {

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)

	orderHandler := CourierHandler{service}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/assignments?courier_id=%s", "a"), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.couriersAssignments(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCourierAssignmentsDbFails(t *testing.T) {

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	e := echo.New()

	repo := mock_courier.NewMockCourierRepository(ctl)
	service := courierService.NewCourierService(repo)
	date, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	courierId := 1

	repo.EXPECT().GetCouriersWithOrdersForDate(date, courierId).Return(nil, errors.New("db is down")).Times(1)

	orderHandler := CourierHandler{service}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/couriers/assignments?courier_id=%d", courierId), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.couriersAssignments(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
