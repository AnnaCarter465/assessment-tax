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
		TotalIncome float64     `json:"totalIncome" validate:"required,number"`
		Wht         float64     `json:"wht" validate:"number"`
		Allowances  []Allowance `json:"allowances" validate:"required,dive,required"`
	}

	Allowance struct {
		AllowanceType string  `json:"allowanceType" validate:"required,lowercase"`
		Amount        float64 `json:"amount" validate:"number"`
	}

	TaxResponse struct {
		Tax float64 `json:"tax"`
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

		tx := tax.NewTax(tax.TaxConfig{
			Rates: []tax.Rate{
				{Percentage: 0, Max: 150_000},
				{Percentage: 0.1, Max: 500_000},
				{Percentage: 0.15, Max: 1_000_000},
				{Percentage: 0.2, Max: 2_000_000},
				{Percentage: 0.35, Max: -1},
			},
			DefaultAllowances: defaultAllowancesMap,
			AllowedAllowances: allowedAllowancesMap,
		}).SetIncome(req.TotalIncome)

		for _, a := range req.Allowances {
			tx.AddAllowance(a.AllowanceType, a.Amount)
		}

		return c.JSON(http.StatusOK, &TaxResponse{
			Tax: tx.CalculateTax(),
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
		e.Logger.Fatal(err)
	}
}
