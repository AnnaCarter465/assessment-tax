package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ResponseMsg struct {
	Message string `json:"message"`
}

func Healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, ResponseMsg{
		Message: "I'm fine, Thank!",
	})
}
