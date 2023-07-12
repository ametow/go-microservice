package courier

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	courierDomain "yandex-team.ru/bstask/internal/courier"
	"yandex-team.ru/bstask/internal/pkg"
	"yandex-team.ru/bstask/internal/pkg/validators"
)

const (
	dateFormat = "2006-01-02"
)

type CourierHandler struct {
	service courierDomain.CourierService
}

func NewHandler(s courierDomain.CourierService) *CourierHandler {
	h := &CourierHandler{s}
	return h
}

func (h *CourierHandler) Init(e *echo.Echo) {
	g := e.Group("/couriers")
	g.GET("", h.getCouriers)
	g.GET("/:courier_id", h.getCourierById)
	g.GET("/meta-info/:courier_id", h.courierMetaInfo)
	g.GET("/assignments", h.couriersAssignments)
	g.POST("", h.createCourier)
}

// e.GET("/:courier_id", getCourierById)
func (h *CourierHandler) getCourierById(ctx echo.Context) error {
	courierIdStr := ctx.Param("courier_id")
	courierId, err := strconv.Atoi(courierIdStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	res, err := h.service.FetchSingleCourier(courierId)
	if err != nil {
		if errors.Is(err, courierDomain.ErrCourierNotFound) {
			return ctx.JSON(http.StatusNotFound, pkg.NotFoundResponse{})
		}
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}
	return ctx.JSON(http.StatusOK, res)
}

// e.GET("/couriers", getCouriers)
func (h *CourierHandler) getCouriers(ctx echo.Context) error {
	limit := ctx.QueryParam("limit")
	if limit == "" {
		limit = "1"
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	offset := ctx.QueryParam("offset")
	if offset == "" {
		offset = "0"
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	res, err := h.service.FetchCouriers(limitInt, offsetInt)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}
	return ctx.JSON(http.StatusOK, res)
}

// e.POST("/couriers", createCourier)
func (h *CourierHandler) createCourier(ctx echo.Context) error {
	in := new(courierDomain.CreateCourierRequest)
	err := ctx.Bind(in)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	err = validateCreateCourierReq(in)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}

	res, err := h.service.CreateNewCouriers(in)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}

	return ctx.JSON(http.StatusOK, res)
}

// e.GET("/couriers/meta-info/:courier_id", courierMetaInfo)
func (h *CourierHandler) courierMetaInfo(ctx echo.Context) error {
	courierIdStr := ctx.Param("courier_id")
	courierId, err := strconv.Atoi(courierIdStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	startDateStr := ctx.QueryParam("startDate")
	startDate, err := time.Parse(dateFormat, startDateStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	endDateStr := ctx.QueryParam("endDate")
	endDate, err := time.Parse(dateFormat, endDateStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}

	response, err := h.service.FetchCourierMetaData(courierId, startDate, endDate)
	if err != nil {
		if errors.Is(err, courierDomain.ErrCourierNotFound) {
			return ctx.JSON(http.StatusNotFound, pkg.NotFoundResponse{})
		}
		return ctx.JSON(http.StatusInternalServerError, pkg.InternalErrorResponse{})
	}

	return ctx.JSON(http.StatusOK, response)
}

// e.GET("/couriers/assignments", couriersAssignments)
func (h *CourierHandler) couriersAssignments(ctx echo.Context) error {
	dateStr := ctx.QueryParam("date")
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		date, _ = time.Parse(dateFormat, time.Now().Format(dateFormat))
	}
	var courierId int
	courierIdStr := ctx.QueryParam("courier_id")
	if courierIdStr != "" {
		courierId, err = strconv.Atoi(courierIdStr)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
		}
	}
	res, err := h.service.FetchCouriersAssignments(date, courierId)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, pkg.BadRequestResponse{})
	}
	return ctx.JSON(http.StatusOK, res)
}

func validateCreateCourierReq(r *courierDomain.CreateCourierRequest) error {
	if len(r.Couriers) == 0 {
		return courierDomain.ErrZeroLengthCouriers
	}
	for _, c := range r.Couriers {
		err := validateCreateCourierDto(&c)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCreateCourierDto(r *courierDomain.CreateCourierDto) error {
	if _, ok := courierDomain.CourierTypes[r.CourierType]; !ok {
		return courierDomain.ErrCourierBadType
	}
	if len(r.Regions) == 0 {
		return courierDomain.ErrCourierBadRegions
	}
	for i := 0; i < len(r.Regions); i++ {
		if r.Regions[i] <= 0 {
			return courierDomain.ErrCourierBadRegions
		}
	}
	if len(r.WorkingHours) == 0 {
		return courierDomain.ErrCourierBadWorkingHours
	}
	err := validators.ValidateHoursSlice(r.WorkingHours)
	return err
}
