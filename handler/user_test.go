package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/AnnaCarter465/assessment-tax/database"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UserDBMock struct {
	mock.Mock
}

func (o *UserDBMock) FindAllDefaultAllowances(ctx context.Context) ([]database.DefaultAllowance, error) {
	args := o.Called(ctx)
	return args.Get(0).([]database.DefaultAllowance), args.Error(1)
}

func (o *UserDBMock) FindAllAllowedAllowances(ctx context.Context) ([]database.AllowedAllowance, error) {
	args := o.Called(ctx)
	return args.Get(0).([]database.AllowedAllowance), args.Error(1)
}

func TestUserCalculateTax(t *testing.T) {
	type TC struct {
		reqbody                      map[string]interface{}
		want                         *TaxResponse
		mockFindAllDefaultAllowances *MockSetting
		mockFindAllAllowedAllowances *MockSetting
		errresp                      *ResponseMsg
	}

	tcs := []TC{
		{
			reqbody: map[string]interface{}{
				"totalIncome": float64(500_000),
				"wht":         float64(0),
				"allowances": []Allowance{
					{AllowanceType: "donation", Amount: 0},
				},
			},
			want: &TaxResponse{
				Tax:       29_000,
				TaxRefund: 0,
				TaxLevel: []TaxLevel{
					{
						Level: "0-150,000",
						Tax:   0,
					},
					{
						Level: "150,001-500,000",
						Tax:   29_000,
					},
					{
						Level: "500,001-1,000,000",
						Tax:   0,
					},
					{
						Level: "1,000,001-2,000,000",
						Tax:   0,
					},
					{
						Level: "2,000,001 ขึ้นไป",
						Tax:   0,
					},
				},
			},
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{
						{AllowanceType: "personal", Amount: 60_000},
					},
					nil,
				},
			},
			mockFindAllAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.AllowedAllowance{
						{AllowanceType: "donation", MaxAmount: 100_000},
						{AllowanceType: "k-receipt", MaxAmount: 50_000},
					},
					nil,
				},
			},
			errresp: nil,
		},
		{
			reqbody: map[string]interface{}{
				"totalIncome": "wrong_input",
			},
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Bad request",
			},
		},
		{
			reqbody: map[string]interface{}{
				"totalIncome": float64(500_000),
				"wht":         float64(-1),
				"allowances":  nil,
			},
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Bad request",
			},
		},
		{
			reqbody: map[string]interface{}{
				"totalIncome": float64(500_000),
				"wht":         float64(500_001),
				"allowances": []Allowance{
					{AllowanceType: "donation", Amount: 0},
				},
			},
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid wht",
			},
		},
		{
			reqbody: map[string]interface{}{
				"totalIncome": float64(500_000),
				"wht":         float64(0),
				"allowances": []Allowance{
					{AllowanceType: "donation", Amount: 0},
				},
			},
			want: nil,
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{},
					errors.New("an error"),
				},
			},
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Internal server error",
			},
		},
		{
			reqbody: map[string]interface{}{
				"totalIncome": float64(500_000),
				"wht":         float64(0),
				"allowances": []Allowance{
					{AllowanceType: "donation", Amount: 0},
				},
			},
			want: nil,
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{
						{AllowanceType: "personal", Amount: 60_000},
					},
					nil,
				},
			},
			mockFindAllAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.AllowedAllowance{},
					errors.New("an error"),
				},
			},
			errresp: &ResponseMsg{
				Message: "Internal server error",
			},
		},
		{
			reqbody: map[string]interface{}{ // exp07
				"totalIncome": float64(500_000),
				"wht":         float64(0),
				"allowances": []Allowance{
					{AllowanceType: "k-receipt", Amount: 200_000},
					{AllowanceType: "donation", Amount: 100_000},
				},
			},
			want: &TaxResponse{
				Tax:       14_000,
				TaxRefund: 0,
				TaxLevel: []TaxLevel{
					{
						Level: "0-150,000",
						Tax:   0,
					},
					{
						Level: "150,001-500,000",
						Tax:   14_000,
					},
					{
						Level: "500,001-1,000,000",
						Tax:   0,
					},
					{
						Level: "1,000,001-2,000,000",
						Tax:   0,
					},
					{
						Level: "2,000,001 ขึ้นไป",
						Tax:   0,
					},
				},
			},
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{
						{AllowanceType: "personal", Amount: 60_000},
					},
					nil,
				},
			},
			mockFindAllAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.AllowedAllowance{
						{AllowanceType: "donation", MaxAmount: 100_000},
						{AllowanceType: "k-receipt", MaxAmount: 50_000},
					},
					nil,
				},
			},
			errresp: nil,
		},
	}

	for i, tc := range tcs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockObj := new(UserDBMock)

			if tc.mockFindAllDefaultAllowances != nil {
				mockObj.On(
					"FindAllDefaultAllowances",
					tc.mockFindAllDefaultAllowances.Args...,
				).Return(tc.mockFindAllDefaultAllowances.Returns...)
			}

			if tc.mockFindAllAllowedAllowances != nil {
				mockObj.On(
					"FindAllAllowedAllowances",
					tc.mockFindAllAllowedAllowances.Args...,
				).Return(tc.mockFindAllAllowedAllowances.Returns...)
			}

			h := NewTaxHandler(validator.New(), mockObj)

			val, _ := json.Marshal(tc.reqbody)

			req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(string(val)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := echo.New()

			goterr := h.CalculateTax(e.NewContext(req, rec))

			assert.NoError(t, goterr)

			if tc.errresp != nil {
				var errresp ResponseMsg

				err := json.Unmarshal([]byte(rec.Body.String()), &errresp)
				assert.NoError(t, err)

				assert.NotEqual(t, http.StatusOK, rec.Code)

				equal := reflect.DeepEqual(*tc.errresp, errresp)

				if !equal {
					assert.Fail(t, fmt.Sprintf("expected %v, \nbut got %v", *tc.errresp, errresp))
				}

				return
			}

			var got TaxResponse

			err := json.Unmarshal([]byte(rec.Body.String()), &got)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)

			equal := reflect.DeepEqual(*tc.want, got)

			if !equal {
				assert.Fail(t, fmt.Sprintf("expected %#v, \nbut got %#v", *tc.want, got))
			}
		})
	}
}

func TestUserCalculateTaxWithCSV(t *testing.T) {
	type TC struct {
		reqbody                      string
		contentType                  string
		want                         *TaxCSVResponse
		mockFindAllDefaultAllowances *MockSetting
		mockFindAllAllowedAllowances *MockSetting
		errresp                      *ResponseMsg
	}

	tcs := []TC{
		{
			reqbody: `
totalIncome,wht,donation
500000,0,0
600000,40000,20000
750000,50000,15000
`,
			contentType: "text/csv",
			want: &TaxCSVResponse{
				Taxes: []TaxCSV{
					{
						TotalIncome: 500000,
						Tax:         29000,
					},
					{
						TotalIncome: 600000,
						Tax:         10000,
					},
					{
						TotalIncome: 750000,
						Tax:         3750,
					},
				},
			},
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{
						{AllowanceType: "personal", Amount: 60_000},
					},
					nil,
				},
			},
			mockFindAllAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.AllowedAllowance{
						{AllowanceType: "donation", MaxAmount: 100_000},
						{AllowanceType: "k-receipt", MaxAmount: 50_000},
					},
					nil,
				},
			},
			errresp: nil,
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,0,0
600000,40000,20000
750000,50000,15000
`,
			contentType:                  "application/json",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Unaceptable content, require CSV content",
			},
		},
		{
			reqbody:                      "",
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Wrong csv content, no content",
			},
		},
		{
			reqbody:                      "totalIncome,wht,donation",
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Wrong csv content, should have more than 1 row due to it is header",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation,k-receipt
500000,0,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Bad request, might not be csv format",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,0,0
600000,40000,20000,0
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Bad request, might not be csv format",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation,k-receipt
500000,0,0,0
600000,40000,20000,0
750000,50000,15000,0`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Wrong csv column length",
			},
		},
		{
			reqbody: `
totalIncome,wht,k-receipt
500000,0,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Wrong csv header",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
aaaa,0,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid income amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,aaaaa,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid wht amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,0,aaaaa
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid donation amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
-1,0,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid income amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,-1,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid wht amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,0,-1
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Invalid donation amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,600000,0
600000,40000,20000
750000,50000,15000`,
			contentType:                  "text/csv",
			want:                         nil,
			mockFindAllDefaultAllowances: nil,
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Income amount should be more than wht amount",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,0,0
600000,40000,20000
750000,50000,15000`,
			contentType: "text/csv",
			want:        nil,
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{},
					errors.New("an error"),
				},
			},
			mockFindAllAllowedAllowances: nil,
			errresp: &ResponseMsg{
				Message: "Internal server error",
			},
		},
		{
			reqbody: `
totalIncome,wht,donation
500000,0,0
600000,40000,20000
750000,50000,15000`,
			contentType: "text/csv",
			want:        nil,
			mockFindAllDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.DefaultAllowance{
						{AllowanceType: "personal", Amount: 60_000},
					},
					nil,
				},
			},
			mockFindAllAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
				},
				Returns: []interface{}{
					[]database.AllowedAllowance{},
					errors.New("an error"),
				},
			},
			errresp: &ResponseMsg{
				Message: "Internal server error",
			},
		},
	}

	for i, tc := range tcs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockObj := new(UserDBMock)

			if tc.mockFindAllDefaultAllowances != nil {
				mockObj.On(
					"FindAllDefaultAllowances",
					tc.mockFindAllDefaultAllowances.Args...,
				).Return(tc.mockFindAllDefaultAllowances.Returns...)
			}

			if tc.mockFindAllAllowedAllowances != nil {
				mockObj.On(
					"FindAllAllowedAllowances",
					tc.mockFindAllAllowedAllowances.Args...,
				).Return(tc.mockFindAllAllowedAllowances.Returns...)
			}

			h := NewTaxHandler(validator.New(), mockObj)

			req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", strings.NewReader(tc.reqbody))
			req.Header.Set("Content-Type", tc.contentType)
			rec := httptest.NewRecorder()

			e := echo.New()

			goterr := h.CalculateTaxWithCSV(e.NewContext(req, rec))

			assert.NoError(t, goterr)

			if tc.errresp != nil {
				var errresp ResponseMsg

				err := json.Unmarshal([]byte(rec.Body.String()), &errresp)
				assert.NoError(t, err)

				assert.NotEqual(t, http.StatusOK, rec.Code)

				equal := reflect.DeepEqual(*tc.errresp, errresp)

				if !equal {
					assert.Fail(t, fmt.Sprintf("expected %v, \nbut got %v", *tc.errresp, errresp))
				}

				return
			}

			var got TaxCSVResponse

			err := json.Unmarshal([]byte(rec.Body.String()), &got)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)

			equal := reflect.DeepEqual(*tc.want, got)

			if !equal {
				assert.Fail(t, fmt.Sprintf("expected %#v, \nbut got %#v", *tc.want, got))
			}
		})
	}
}
