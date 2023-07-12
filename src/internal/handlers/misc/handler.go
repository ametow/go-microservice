package misc

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type handler struct{}

func NewHandler(e *echo.Echo) {
	h := &handler{}
	e.GET("/ping", h.ping)
}

// e.GET("/ping", ping)
func (handler) ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}
