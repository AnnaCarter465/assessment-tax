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
