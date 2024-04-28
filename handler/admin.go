package handler

import (
	"log"
	"net/http"

	"github.com/AnnaCarter465/assessment-tax/database"
	"github.com/labstack/echo/v4"
)

type AdminTaxRequest struct {
	Amount float64 `json:"amount" validate:"required,number,gte=0"`
}

type AdminHandler struct {
	db *database.DB
}

func NewAdminHandler(db *database.DB) *AdminHandler {
	return &AdminHandler{db}
}

func (a *AdminHandler) UpdatePesonal(c echo.Context) error {
	var req AdminTaxRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "bad request",
		})
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if req.Amount < 10_000 || req.Amount > 100_000 {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "invalid amount",
		})
	}

	defaultAllowance, err := a.db.UpdateAmountDefaultAllowances(c.Request().Context(), "personal", req.Amount)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Failed to update personal amount",
		})
	}

	return c.JSON(http.StatusOK, map[string]float64{
		"personalDeduction": defaultAllowance.Amount,
	})
}
