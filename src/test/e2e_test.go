package test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
	"yandex-team.ru/bstask/internal/handlers/courier"
	"yandex-team.ru/bstask/internal/handlers/misc"
	"yandex-team.ru/bstask/internal/handlers/order"
	orderDomain "yandex-team.ru/bstask/internal/order"
	courierRepo "yandex-team.ru/bstask/internal/pkg/repository/courier"
	orderRepo "yandex-team.ru/bstask/internal/pkg/repository/order"
	courierService "yandex-team.ru/bstask/internal/usecase/courier"
	orderService "yandex-team.ru/bstask/internal/usecase/order"
)

func TestE2E(t *testing.T) {
	const prefix = "../"
	db, err := PrepareTestDatabase(prefix)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err.Error())
	}

	app := echo.New()

	courierRepo := courierRepo.NewRepo(db)
	cService := courierService.NewCourierService(courierRepo)
	courierHandler := courier.NewHandler(cService)
	courierHandler.Init(app)

	orderRepo := orderRepo.NewRepo(db)
	oService := orderService.NewOrderService(&orderRepo)
	orderHandler := order.NewHandler(oService)
	orderHandler.Init(app)

	misc.NewHandler(app)

	COURIER_URL := "/couriers"
	ORDER_URL := "/orders"

	Convey("Courier tests", t, func() {
		Convey("Create couriers", func() {
			input := `{"couriers":[{"courier_type":"AUTO","regions":[5,15],"working_hours":["16:30-18:45"]},{"courier_type":"AUTO","regions":[5,11],"working_hours":["16:00-17:00"]},{ "courier_type":"BIKE","regions":[5,44],"working_hours":["16:00-16:20"]},{"courier_type":"BIKE","regions":[5,44],"working_hours":["08:00-18:00","20:00-22:00"]},{"courier_type":"BIKE","regions":[6,16],"working_hours":["16:00-18:00","20:00-22:00"]},{"courier_type":"FOOT","regions":[5,1],"working_hours":["16:00-18:00","20:00-22:00"]},{"courier_type":"FOOT","regions":[5,1],"working_hours":["14:00-14:00"]}]}`
			output := `{"couriers":[{"courier_id":1,"courier_type":"AUTO","regions":[5,15],"working_hours":["16:30-18:45"]},{"courier_id":2,"courier_type":"AUTO","regions":[5,11],"working_hours":["16:00-17:00"]},{"courier_id":3,"courier_type":"BIKE","regions":[5,44],"working_hours":["16:00-16:20"]},{"courier_id":4,"courier_type":"BIKE","regions":[5,44],"working_hours":["08:00-18:00","20:00-22:00"]},{"courier_id":5,"courier_type":"BIKE","regions":[6,16],"working_hours":["16:00-18:00","20:00-22:00"]},{"courier_id":6,"courier_type":"FOOT","regions":[5,1],"working_hours":["16:00-18:00","20:00-22:00"]},{"courier_id":7,"courier_type":"FOOT","regions":[5,1],"working_hours":["14:00-14:00"]}]}`
			req := httptest.NewRequest(echo.POST, COURIER_URL, strings.NewReader(input))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			body := w.Body
			Convey("So response is", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(body.String(), ShouldEqualJSON, output)
			})
		})
		Convey("Fail to create couriers", func() {
			req := httptest.NewRequest(echo.POST, COURIER_URL, strings.NewReader(``))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			Convey("So response is", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("Given params", func() {
			const (
				limit  = 10
				offset = 0
			)

			Convey("Get couriers assignments", func() {
				req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s/assignments", COURIER_URL), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				Convey("Then should be Ok", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
			Convey("Get couriers", func() {
				req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s?limit=%d&offset=%d", COURIER_URL, limit, offset), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				Convey("Then should be Ok", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})
		Convey("Given courier id", func() {
			const (
				courierId = 1
				output    = `{"courier_id":1,"courier_type":"AUTO","regions":[5,15],"working_hours":["16:30-18:45"]}`
			)
			Convey("Get courier meta-data of id 1", func() {
				req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s/meta-info/%d?startDate=2023-01-02&endDate=2023-03-02", COURIER_URL, courierId), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				Convey("Then should be Ok", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
			Convey("Get courier with id = 1", func() {
				req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s/%d", COURIER_URL, courierId), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				Convey("Then should be Ok", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqualJSON, output)
				})
			})
			Convey("Get courier with id = 99", func() {
				req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s/%d", COURIER_URL, 99), nil)
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				Convey("Then should be NotFound", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})
	})

	Convey("Order test", t, func() {
		const (
			limit  = 10
			offset = 0
		)

		Convey("Create orders", func() {
			input := `{ "orders": [ { "cost": 120, "delivery_hours": [ "16:05-16:30", "13:00-15:30" ], "regions": 5, "weight": 45.0 }, { "cost": 99, "delivery_hours": [ "16:10-16:20", "13:00-15:30" ], "regions": 5, "weight": 35.0 }, { "cost": 150, "delivery_hours": [ "16:15-17:00", "13:00-15:30" ], "regions": 5, "weight": 10.0 }, { "cost": 250, "delivery_hours": [ "16:30-16:50", "13:00-15:30" ], "regions": 5, "weight": 9.0 }, { "cost": 350, "delivery_hours": [ "17:00-18:00", "13:00-15:30" ], "regions": 5, "weight": 5.0 }, { "cost": 150, "delivery_hours": [ "17:20-18:00", "13:00-15:30" ], "regions": 6, "weight": 4.0 }, { "cost": 150, "delivery_hours": [ "17:20-17:30", "13:00-15:30" ], "regions": 5, "weight": 2.0 } ] }`
			output := `[{"cost":120,"delivery_hours":["16:05-16:30","13:00-15:30"],"order_id":1,"regions":5,"weight":45},{"cost":99,"delivery_hours":["16:10-16:20","13:00-15:30"],"order_id":2,"regions":5,"weight":35},{"cost":150,"delivery_hours":["16:15-17:00","13:00-15:30"],"order_id":3,"regions":5,"weight":10},{"cost":250,"delivery_hours":["16:30-16:50","13:00-15:30"],"order_id":4,"regions":5,"weight":9},{"cost":350,"delivery_hours":["17:00-18:00","13:00-15:30"],"order_id":5,"regions":5,"weight":5},{"cost":150,"delivery_hours":["17:20-18:00","13:00-15:30"],"order_id":6,"regions":6,"weight":4},{"cost":150,"delivery_hours":["17:20-17:30","13:00-15:30"],"order_id":7,"regions":5,"weight":2}]`
			req := httptest.NewRequest(echo.POST, ORDER_URL, strings.NewReader(input))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			body := w.Body
			Convey("So response is", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(body.String(), ShouldEqualJSON, output)
			})
		})

		Convey("Get orders", func() {
			req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s?limit=%d&offset=%d", ORDER_URL, limit, offset), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			Convey("Then should be Ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
			Convey("Also should be parsable", func() {
				expectedOutput := []orderDomain.OrderDto{}
				So(json.Unmarshal(w.Body.Bytes(), &expectedOutput), ShouldEqual, nil)
			})
			shouldBeLength := 7
			Convey("Should be of length", func() {
				expectedOutput := []orderDomain.OrderDto{}
				json.Unmarshal(w.Body.Bytes(), &expectedOutput)
				So(len(expectedOutput), ShouldEqual, shouldBeLength)
			})
		})

		Convey("Get order with id = 1", func() {
			req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s/%d", ORDER_URL, 1), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			output := `{"cost":120,"delivery_hours":["16:05-16:30","13:00-15:30"],"order_id":1,"regions":5,"weight":45}`
			Convey("Then should be Ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Body.String(), ShouldEqualJSON, output)
			})
		})
		Convey("Get order with id = 99 should fail", func() {
			req := httptest.NewRequest(echo.GET, fmt.Sprintf("%s/%d", ORDER_URL, 99), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			Convey("Then should be NotFound", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("Assign orders", func() {
			req := httptest.NewRequest(echo.POST, fmt.Sprintf("%s/assign", ORDER_URL), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			Convey("Then should be Ok", func() {
				So(w.Code, ShouldEqual, http.StatusCreated)
			})
		})
		Convey("Complete orders", func() {
			input := `{ "complete_info": [ { "complete_time": "2023-04-01T10:08:11+05:00", "courier_id": 2, "order_id": 2 } ] }`
			output := `[{"completed_time":"2023-04-01T10:08:11+05:00","cost":99,"delivery_hours":["16:10-16:20","13:00-15:30"],"order_id":2,"regions":5,"weight":35}]`
			req := httptest.NewRequest(echo.POST, fmt.Sprintf("%s/complete", ORDER_URL), strings.NewReader(input))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			Convey("Then should be Ok", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
			Convey("Also should be equal to expected", func() {
				So(w.Body.String(), ShouldEqualJSON, output)
			})
		})
	})
}
