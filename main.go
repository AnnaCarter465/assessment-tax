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
	"github.com/AnnaCarter465/assessment-tax/handler"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	ResponseMsg struct {
		Message string `json:"message"`
	}

	AdminTaxRequest struct {
		Amount float64 `json:"amount" validate:"required,number,gte=0"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, ResponseMsg{
			Message: err.Error(),
		})
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

	e.GET("/", handler.Healthcheck)
	e.POST("/tax/calculations", handler.NewTaxHandler(db).CalculateTax)

	// admin -----------------------------------------------------------------------------
	am := e.Group("/admin")
	am.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD") {
			return true, nil
		}
		return false, nil
	}))

	am.POST("/deductions/personal", handler.NewAdminHandler(db).UpdatePesonal)

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
