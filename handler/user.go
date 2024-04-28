package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/AnnaCarter465/assessment-tax/database"
	"github.com/AnnaCarter465/assessment-tax/tax"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type (
	TaxRequest struct {
		TotalIncome float64     `json:"totalIncome" validate:"required,number,gte=0"`
		Wht         float64     `json:"wht" validate:"number,gte=0"`
		Allowances  []Allowance `json:"allowances" validate:"required,dive"`
	}

	Allowance struct {
		AllowanceType string  `json:"allowanceType" validate:"required,lowercase"`
		Amount        float64 `json:"amount" validate:"number,gte=0"`
	}

	TaxResponse struct {
		Tax       float64    `json:"tax"`
		TaxRefund float64    `json:"taxRefund"`
		TaxLevel  []TaxLevel `json:"taxLevel"`
	}

	TaxLevel struct {
		Level string  `json:"level"`
		Tax   float64 `json:"tax"`
	}
)

type IDB interface {
	FindAllDefaultAllowances(ctx context.Context) ([]database.DefaultAllowance, error)
	FindAllAllowedAllowances(ctx context.Context) ([]database.AllowedAllowance, error)
}

type TaxHandler struct {
	vl *validator.Validate
	db IDB
}

func NewTaxHandler(vl *validator.Validate, db IDB) *TaxHandler {
	return &TaxHandler{vl, db}
}

func (t *TaxHandler) CalculateTax(c echo.Context) error {
	var req TaxRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Bad request",
		})
	}

	if err := t.vl.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Bad request",
		})
	}

	if req.TotalIncome < req.Wht {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Invalid wht",
		})
	}

	defaultAllowances, err := t.db.FindAllDefaultAllowances(c.Request().Context())
	if err != nil {
		log.Println("Failed to find all default allowaces:", err)
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Internal server error",
		})
	}

	defaultAllowancesMap := make(tax.Allowances)

	for _, defaultAllowance := range defaultAllowances {
		defaultAllowancesMap[defaultAllowance.AllowanceType] = defaultAllowance.Amount
	}

	allowedAllowances, err := t.db.FindAllAllowedAllowances(c.Request().Context())
	if err != nil {
		log.Println("Failed to find all allowed allowaces:", err)
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Internal server error",
		})
	}

	allowedAllowancesMap := make(tax.Allowances)

	for _, allowedAllowance := range allowedAllowances {
		allowedAllowancesMap[allowedAllowance.AllowanceType] = allowedAllowance.MaxAmount
	}

	rates := []tax.Rate{
		{Percentage: 0, Max: 150_000, Label: "0-150,000"},
		{Percentage: 0.1, Max: 500_000, Label: "150,001-500,000"},
		{Percentage: 0.15, Max: 1_000_000, Label: "500,001-1,000,000"},
		{Percentage: 0.2, Max: 2_000_000, Label: "1,000,001-2,000,000"},
		{Percentage: 0.35, Max: -1, Label: "2,000,001 ขึ้นไป"},
	}

	tx := tax.NewTax(tax.TaxConfig{
		Rates:             rates,
		DefaultAllowances: defaultAllowancesMap,
		AllowedAllowances: allowedAllowancesMap,
	}).SetIncome(req.TotalIncome).SetWht(req.Wht)

	for _, a := range req.Allowances {
		tx.AddAllowance(a.AllowanceType, a.Amount)
	}

	summary := tx.CalculateTaxSummary()

	var levels []TaxLevel

	for _, l := range summary.TaxStatements {
		levels = append(levels, TaxLevel{
			Level: l.Rate.Label,
			Tax:   l.Tax,
		})
	}

	return c.JSON(http.StatusOK, &TaxResponse{
		Tax:       summary.Tax,
		TaxRefund: summary.Refund,
		TaxLevel:  levels,
	})
}
