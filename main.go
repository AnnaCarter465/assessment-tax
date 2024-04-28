package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/AnnaCarter465/assessment-tax/database"
	"github.com/AnnaCarter465/assessment-tax/tax"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type (
	TaxRequest struct {
		TotalIncome float64     `json:"totalIncome" validate:"required,number,gte=0"`
		Wht         float64     `json:"wht" validate:"number,gte=0"`
		Allowances  []Allowance `json:"allowances" validate:"required,dive,required"`
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

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	if len(strings.TrimSpace(dbURL)) == 0 {
		log.Fatal("Missing an env variable `DATABASE_URL`")
	}

	db, err := database.NewDB(dbURL)
	if err != nil {
		log.Fatal("Cannot connection to database", err)
	}

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	e.POST("/tax/calculations", func(c echo.Context) error {
		var req TaxRequest

		if err := c.Bind(&req); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		if err = c.Validate(req); err != nil {
			return err
		}

		if req.TotalIncome < req.Wht {
			return c.String(http.StatusBadRequest, "wht must less than total income")
		}

		defaultAllowances, err := db.FindAllDefaultAllowances(c.Request().Context())
		if err != nil {
			log.Println("Failed to find all allowed allownaces", err)
			return err
		}

		defaultAllowancesMap := make(tax.Allowances)

		for _, defaultAllowance := range defaultAllowances {
			defaultAllowancesMap[defaultAllowance.AllowanceType] = defaultAllowance.Amount
		}

		allowedAllowances, err := db.FindAllAllowedAllowances(c.Request().Context())
		if err != nil {
			log.Println("Failed to find all allowed allownaces", err)
			return err
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
	})

	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown

	log.Println("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(e.Start(":8000"))
		e.Logger.Fatal(err)
	}
}
