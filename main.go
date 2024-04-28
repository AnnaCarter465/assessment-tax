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

type TaxRequest struct {
	TotalIncome float64     `json:"totalIncome" validate:"required,number,gte=0"`
	Wht         float64     `json:"wht" validate:"number,gte=0"`
	Allowances  []Allowance `json:"allowances" validate:"required,dive,required"`
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

	vl := validator.New()

	e := echo.New()

	e.GET("/", handler.Healthcheck)

	// user ------------------------------------------------------------------------------
	u := e.Group("/tax")
	u.POST("/calculations", handler.NewTaxHandler(vl, db).CalculateTax)
	u.POST("/calculations/upload-csv", handler.NewTaxHandler(vl, db).CalculateTaxWithCSV)

	// admin -----------------------------------------------------------------------------
	am := e.Group("/admin")
	am.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD") {
			return true, nil
		}
		return false, nil
	}))

	am.POST("/deductions/personal", handler.NewAdminHandler(vl, db).UpdatePesonal)

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
