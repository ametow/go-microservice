package order

import (
	"database/sql"
	"errors"
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
	"yandex-team.ru/bstask/internal/order"
	"yandex-team.ru/bstask/internal/pkg"
	mock_order "yandex-team.ru/bstask/internal/pkg/repository/order/mocks"
	orderService "yandex-team.ru/bstask/internal/usecase/order"
)

func TestInit(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := orderService.NewOrderService(repo)
	h := NewHandler(service)
	h.Init(e)
}
func TestNewHandler(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := orderService.NewOrderService(repo)
	NewHandler(service)
}

func TestGetOrderSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_order.NewMockOrderRepository(ctl)
	service := orderService.NewOrderService(repo)

	repo.EXPECT().GetOrderByID(47).Return(&order.Order{ID: 5}, nil).Times(1)

	orderHandler := OrderHandler{service}

	rec := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	c := e.NewContext(req, rec)

	c.SetParamNames("order_id")
	c.SetParamValues("47")

	require.NoError(t, orderHandler.getOrder(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetOrderInvalidParamFails(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders/47", nil)
	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.getOrder(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
func TestGetOrderDbDown(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	repo.EXPECT().GetOrderByID(47).Return(nil, errors.New("db is down")).Times(1)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders/", nil)

	c := e.NewContext(req, rec)
	c.SetParamNames("order_id")
	c.SetParamValues("47")

	require.NoError(t, orderHandler.getOrder(c))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetOrdersSuccess(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_order.NewMockOrderRepository(ctl)

	repo.EXPECT().GetOrders(10, 0).Return([]order.Order{
		{
			ID:     1,
			Cost:   120,
			Weight: 2.3,
			Region: 23,
			DeliveryHours: []order.OrderDeliveryHours{
				{
					Starts: pkg.TIME{},
					Ends:   pkg.TIME{},
				},
			},
		},
	}, nil).Times(1)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}

	rec := httptest.NewRecorder()

	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "0")
	req := httptest.NewRequest(http.MethodGet, "/orders?"+q.Encode(), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.getOrders(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "[{\"cost\":120,\"delivery_hours\":[\"00:00-00:00\"],\"order_id\":1,\"regions\":23,\"weight\":2.3}]\n", rec.Body.String())
}

func TestGetOrdersDbDown(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()
	repo := mock_order.NewMockOrderRepository(ctl)
	repo.EXPECT().GetOrders(10, 0).Return(nil, errors.New("db is down")).Times(1)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}
	rec := httptest.NewRecorder()
	q := make(url.Values)
	q.Set("limit", "10")
	q.Set("offset", "0")
	req := httptest.NewRequest(http.MethodGet, "/orders?"+q.Encode(), nil)

	c := e.NewContext(req, rec)

	require.NoError(t, orderHandler.getOrders(c))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
func TestGetOrdersInvalidParamsFail(t *testing.T) {
	ctl := gomock.NewController(t)
	e := echo.New()
	defer ctl.Finish()

	repo := mock_order.NewMockOrderRepository(ctl)
	repo.EXPECT().GetOrders(1, 0).Return([]order.Order{}, nil).Times(1)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}

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
			req := httptest.NewRequest(http.MethodGet, "/orders?"+tCase.input, nil)
			c := e.NewContext(req, rec)
			require.NoError(t, orderHandler.getOrders(c))
			require.Equal(t, tCase.expect, rec.Code)
		})
	}

}

var (
	createOrderJson       = `{"orders":[{"cost":120,"delivery_hours":["01:00-11:00","13:00-15:30"],"regions":12,"weight":4.2}]}`
	expectedResFromCreate = "[{\"cost\":120,\"delivery_hours\":[\"01:00-11:00\",\"13:00-15:30\"],\"order_id\":1,\"regions\":12,\"weight\":4.2}]\n"
)

func TestCreateOrder(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	e := echo.New()

	repo := mock_order.NewMockOrderRepository(ctl)

	input := order.CreateOrderDto{
		Weight:        4.2,
		Cost:          120,
		Regions:       12,
		DeliveryHours: []string{"01:00-11:00", "13:00-15:30"},
	}
	repo.EXPECT().CreateOrder(input).Return(uint(1), nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(createOrderJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	c := e.NewContext(req, rec)
	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}

	require.NoError(t, orderHandler.createOrder(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, expectedResFromCreate, rec.Body.String())
}

func TestCreateOrderFail(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	e := echo.New()

	repo := mock_order.NewMockOrderRepository(ctl)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}
	input := order.CreateOrderDto{
		Cost:          120,
		Weight:        4.2,
		Regions:       12,
		DeliveryHours: []string{"01:00-11:00", "13:00-15:30"},
	}
	repo.EXPECT().CreateOrder(input).Return(uint(0), errors.New("db is down")).Times(1)

	tcases := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "invalid_input",
			input:  `{"orders":["cost":120,"delivery_hours":["01:00-11:00","13:00-15:30"],"regions":12,"weight":4.2}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_time",
			input:  `{"orders":[{"cost":120,"delivery_hours":["01:00-11:00","13:00-5:30"],"regions":12,"weight":4.2}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_cost",
			input:  `{"orders":[{"cost":-120,"delivery_hours":["01:00-11:00","13:00-15:30"],"regions":12,"weight":4.2}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "db is down case",
			input:  `{"orders":[{"cost":120,"delivery_hours":["01:00-11:00","13:00-15:30"],"regions":12,"weight":4.2}]}`,
			expect: http.StatusInternalServerError,
		},
	}
	for _, tCase := range tcases {
		t.Run(tCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tCase.input))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			require.NoError(t, orderHandler.createOrder(c))
			require.Equal(t, tCase.expect, rec.Code)
		})
	}
}

var (
	completeOrderJson       = `{"complete_info": [{"order_id":1,"courier_id":1,"complete_time":"2023-04-01T10:08:11+05:00"}]}`
	expectedResFromComplete = "[{\"cost\":120,\"delivery_hours\":[\"00:00-00:00\"],\"order_id\":1,\"regions\":12,\"weight\":2.3,\"completed_time\":\"2023-04-01T10:08:11+05:00\"}]\n"
)

func TestCompleteOrder(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	e := echo.New()

	repo := mock_order.NewMockOrderRepository(ctl)

	completeOrderDtoInput := order.CompleteOrder{
		CourierId:    1,
		OrderId:      1,
		CompleteTime: "2023-04-01T10:08:11+05:00",
	}
	expTime, _ := time.Parse(time.RFC3339, "2023-04-01T10:08:11+05:00")
	expTimeNull := sql.NullTime{}
	err := expTimeNull.Scan(expTime)
	if err != nil {
		log.Println(err)
	}
	expected := order.Order{
		ID:            1,
		Cost:          120,
		Weight:        2.3,
		Region:        12,
		CompletedTime: expTimeNull,
		DeliveryHours: []order.OrderDeliveryHours{
			{
				Starts: pkg.TIME{},
				Ends:   pkg.TIME{},
			},
		},
	}

	repo.EXPECT().CompleteOrder(completeOrderDtoInput).Return(&expected, nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/orders/complete", strings.NewReader(completeOrderJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	c := e.NewContext(req, rec)
	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}

	require.NoError(t, orderHandler.completeOrder(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, expectedResFromComplete, rec.Body.String())
}

func TestCompleteOrderInvalidParamsFail(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	e := echo.New()

	repo := mock_order.NewMockOrderRepository(ctl)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}
	input := order.CompleteOrder{
		CourierId:    1,
		OrderId:      1,
		CompleteTime: "2023-04-07T01:25:22.150Z",
	}
	repo.EXPECT().CompleteOrder(input).Return(nil, errors.New("db is down")).Times(1)

	tcases := []struct {
		name   string
		input  string
		expect int
	}{
		{
			name:   "invalid_json",
			input:  `{"complete_info":["complete_time":"2023-04-07T01:25:22.150Z","courier_id":1,"order_id":1}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_courier_id",
			input:  `{"complete_info":[]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "invalid_time",
			input:  `{"complete_info":[{"complete_time":"2023","courier_id":1,"order_id":1}]}`,
			expect: http.StatusBadRequest,
		},
		{
			name:   "db is down case",
			input:  `{"complete_info":[{"complete_time":"2023-04-07T01:25:22.150Z","courier_id":1,"order_id":1}]}`,
			expect: http.StatusInternalServerError,
		},
	}
	for _, tCase := range tcases {
		t.Run(tCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tCase.input))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			require.NoError(t, orderHandler.completeOrder(c))
			require.Equal(t, tCase.expect, rec.Code)
		})
	}
}

func TestOrdersAssign(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	e := echo.New()

	repo := mock_order.NewMockOrderRepository(ctl)

	service := orderService.NewOrderService(repo)
	orderHandler := OrderHandler{service}

	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	repo.EXPECT().GetUnassignedOrders().Return(nil, nil).Times(1)
	repo.EXPECT().GetFreeCouriers(today).Return([]courier.Courier{
		{
			ID:   1,
			Type: "FOOT",
			Regions: []courier.CourierRegions{
				{
					Number: 21,
				},
			},
			WorkingHours: []courier.CourierWorkingHours{{Starts: pkg.TIME{}, Ends: pkg.TIME{}}},
		},
	}, nil).Times(1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders/assign", nil)

	c := e.NewContext(req, rec)
	require.NoError(t, orderHandler.ordersAssign(c))
	require.Equal(t, http.StatusCreated, rec.Code)
}
