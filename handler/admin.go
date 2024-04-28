package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/AnnaCarter465/assessment-tax/database"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type AdminTaxRequest struct {
	Amount float64 `json:"amount" validate:"required,number,gte=0"`
}

type AdminIDB interface {
	UpdateAmountDefaultAllowances(ctx context.Context, allowanceType string, amount float64) (database.DefaultAllowance, error)
}

type AdminHandler struct {
	vl *validator.Validate
	db AdminIDB
}

func NewAdminHandler(vl *validator.Validate, db AdminIDB) *AdminHandler {
	return &AdminHandler{vl, db}
}

func (a *AdminHandler) UpdatePesonal(c echo.Context) error {
	var req AdminTaxRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Bad request",
		})
	}

	if err := a.vl.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Bad request",
		})
	}

	if req.Amount < 10_000 || req.Amount > 100_000 {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Invalid amount",
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
