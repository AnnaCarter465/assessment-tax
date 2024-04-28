package handler

import (
	"context"
	"encoding/csv"
	"log"
	"net/http"
	"strconv"

	"github.com/AnnaCarter465/assessment-tax/database"
	"github.com/AnnaCarter465/assessment-tax/tax"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type TaxRequest struct {
	TotalIncome float64     `json:"totalIncome" validate:"required,number,gte=0"`
	Wht         float64     `json:"wht" validate:"number,gte=0"`
	Allowances  []Allowance `json:"allowances" validate:"required,dive"`
}

type Allowance struct {
	AllowanceType string  `json:"allowanceType" validate:"required,lowercase"`
	Amount        float64 `json:"amount" validate:"number,gte=0"`
}

type TaxResponse struct {
	Tax       float64    `json:"tax"`
	TaxRefund float64    `json:"taxRefund"`
	TaxLevel  []TaxLevel `json:"taxLevel"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type TaxCSV struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
}

type TaxCSVResponse struct {
	Taxes []TaxCSV `json:"taxes"`
}

var rates = []tax.Rate{
	{Percentage: 0, Max: 150_000, Label: "0-150,000"},
	{Percentage: 0.1, Max: 500_000, Label: "150,001-500,000"},
	{Percentage: 0.15, Max: 1_000_000, Label: "500,001-1,000,000"},
	{Percentage: 0.2, Max: 2_000_000, Label: "1,000,001-2,000,000"},
	{Percentage: 0.35, Max: -1, Label: "2,000,001 ขึ้นไป"},
}

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

func (t *TaxHandler) getDefaultAllowancesMap(ctx context.Context) (tax.Allowances, error) {
	defaultAllowances, err := t.db.FindAllDefaultAllowances(ctx)
	if err != nil {
		log.Println("Failed to find all default allowaces:", err)
		return nil, err
	}

	defaultAllowancesMap := make(tax.Allowances)

	for _, defaultAllowance := range defaultAllowances {
		defaultAllowancesMap[defaultAllowance.AllowanceType] = defaultAllowance.Amount
	}

	return defaultAllowancesMap, nil
}

func (t *TaxHandler) getAllowedAllowancesMap(ctx context.Context) (tax.Allowances, error) {
	allowedAllowances, err := t.db.FindAllAllowedAllowances(ctx)
	if err != nil {
		log.Println("Failed to find all allowed allowaces:", err)
		return nil, err
	}

	allowedAllowancesMap := make(tax.Allowances)

	for _, allowedAllowance := range allowedAllowances {
		allowedAllowancesMap[allowedAllowance.AllowanceType] = allowedAllowance.MaxAmount
	}

	return allowedAllowancesMap, nil
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

	defaultAllowancesMap, err := t.getDefaultAllowancesMap(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Internal server error",
		})
	}

	allowedAllowancesMap, err := t.getAllowedAllowancesMap(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Internal server error",
		})
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

func (t *TaxHandler) CalculateTaxWithCSV(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") != "text/csv" {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Unaceptable content, require CSV content",
		})
	}

	rows, err := csv.NewReader(c.Request().Body).ReadAll()
	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Bad request, might not be csv format",
		})
	}

	if len(rows) == 0 {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Wrong csv content, no content",
		})
	}

	if len(rows) == 1 {
		return c.JSON(http.StatusBadRequest, ResponseMsg{
			Message: "Wrong csv content, should have more than 1 row due to it is header",
		})
	}

	var datasets [][]float64

	// vaildation
	for i, row := range rows {
		if len(row) != 3 {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Wrong csv column length",
			})
		}

		if i == 0 {
			badcsvformat := row[0] != "totalIncome" ||
				row[1] != "wht" ||
				row[2] != "donation"

			if badcsvformat {
				return c.JSON(http.StatusBadRequest, ResponseMsg{
					Message: "Wrong csv header",
				})
			}

			continue
		}

		income, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Invalid income amount",
			})
		}

		wht, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Invalid wht amount",
			})
		}

		donation, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Invalid donation amount",
			})
		}

		if income < 0 {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Invalid income amount",
			})
		}

		if wht < 0 {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Invalid wht amount",
			})
		}

		if donation < 0 {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Invalid donation amount",
			})
		}

		if income < wht {
			return c.JSON(http.StatusBadRequest, ResponseMsg{
				Message: "Income amount should be more than wht amount",
			})
		}

		datasets = append(datasets, []float64{income, wht, donation})
	}

	defaultAllowancesMap, err := t.getDefaultAllowancesMap(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Internal server error",
		})
	}

	allowedAllowancesMap, err := t.getAllowedAllowancesMap(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseMsg{
			Message: "Internal server error",
		})
	}

	var taxes []TaxCSV

	for _, d := range datasets {
		tx := tax.NewTax(tax.TaxConfig{
			Rates:             rates,
			DefaultAllowances: defaultAllowancesMap,
			AllowedAllowances: allowedAllowancesMap,
		})

		summary := tx.
			SetIncome(d[0]).
			SetWht(d[1]).
			AddAllowance("donation", d[2]).
			CalculateTaxSummary()

		taxes = append(taxes, TaxCSV{
			TotalIncome: d[0],
			Tax:         summary.Tax,
		})
	}

	return c.JSON(http.StatusOK, &TaxCSVResponse{
		Taxes: taxes,
	})
}
