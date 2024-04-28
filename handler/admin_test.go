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

type AdminDBMock struct {
	mock.Mock
}

func (o *AdminDBMock) UpdateAmountDefaultAllowances(ctx context.Context, allowanceType string, amount float64) (database.DefaultAllowance, error) {
	args := o.Called(ctx, allowanceType, amount)
	return args.Get(0).(database.DefaultAllowance), args.Error(1)
}

func (o *AdminDBMock) UpdateAmountAllowedAllowances(ctx context.Context, allowanceType string, amount float64) (database.AllowedAllowance, error) {
	args := o.Called(ctx, allowanceType, amount)
	return args.Get(0).(database.AllowedAllowance), args.Error(1)
}

type MockSetting struct {
	Args    []interface{}
	Returns []interface{}
}

func TestAdminUpdatePesonal(t *testing.T) {
	type TC struct {
		reqbody                           map[string]interface{}
		want                              map[string]float64
		mockUpdateAmountDefaultAllowances *MockSetting
		errresp                           *ResponseMsg
	}

	tcs := []TC{
		{
			reqbody: map[string]interface{}{
				"amount": 70_000,
			},
			mockUpdateAmountDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
					"personal",
					float64(70_000),
				},
				Returns: []interface{}{
					database.DefaultAllowance{AllowanceType: "personal", Amount: 70_000},
					nil,
				},
			},
			want: map[string]float64{
				"personalDeduction": 70_000,
			},
			errresp: nil,
		},
		{
			reqbody: map[string]interface{}{
				"amount": "wrong_amount",
			},
			mockUpdateAmountDefaultAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Bad request",
			},
		},
		{
			reqbody:                           nil,
			mockUpdateAmountDefaultAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Bad request",
			},
		},
		{
			reqbody: map[string]interface{}{
				"amount": 9_999,
			},
			mockUpdateAmountDefaultAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Invalid amount",
			},
		},
		{
			reqbody: map[string]interface{}{
				"amount": 100_001,
			},
			mockUpdateAmountDefaultAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Invalid amount",
			},
		},
		{
			reqbody: map[string]interface{}{
				"amount": 70_000,
			},
			mockUpdateAmountDefaultAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
					"personal",
					float64(70_000),
				},
				Returns: []interface{}{
					database.DefaultAllowance{},
					errors.New("an error"),
				},
			},
			want: nil,
			errresp: &ResponseMsg{
				Message: "Failed to update personal amount",
			},
		},
	}

	for i, tc := range tcs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dbmock := new(AdminDBMock)

			if tc.mockUpdateAmountDefaultAllowances != nil {
				dbmock.On(
					"UpdateAmountDefaultAllowances",
					tc.mockUpdateAmountDefaultAllowances.Args...,
				).Return(tc.mockUpdateAmountDefaultAllowances.Returns...)
			}

			h := NewAdminHandler(validator.New(), dbmock)

			val, _ := json.Marshal(tc.reqbody)

			req := httptest.NewRequest(http.MethodPost, "/admin/deductions/personal", strings.NewReader(string(val)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := echo.New()

			goterr := h.UpdatePesonal(e.NewContext(req, rec))

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

			var got map[string]float64

			err := json.Unmarshal([]byte(rec.Body.String()), &got)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)

			equal := reflect.DeepEqual(tc.want, got)

			if !equal {
				assert.Fail(t, fmt.Sprintf("expected %v, \nbut got %v", tc.want, got))
			}
		})
	}
}

func TestAdminUpdateKReciept(t *testing.T) {
	type TC struct {
		reqbody                           map[string]interface{}
		want                              map[string]float64
		mockUpdateAmountAllowedAllowances *MockSetting
		errresp                           *ResponseMsg
	}

	tcs := []TC{
		{
			reqbody: map[string]interface{}{
				"amount": 70_000,
			},
			mockUpdateAmountAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
					"k-receipt",
					float64(70_000),
				},
				Returns: []interface{}{
					database.AllowedAllowance{AllowanceType: "k-receipt", MaxAmount: 70_000},
					nil,
				},
			},
			want: map[string]float64{
				"kReceipt": 70_000,
			},
			errresp: nil,
		},
		{
			reqbody: map[string]interface{}{
				"amount": "wrong_amount",
			},
			mockUpdateAmountAllowedAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Bad request",
			},
		},
		{
			reqbody:                           nil,
			mockUpdateAmountAllowedAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Bad request",
			},
		},
		{
			reqbody: map[string]interface{}{
				"amount": 100_001,
			},
			mockUpdateAmountAllowedAllowances: nil,
			want:                              nil,
			errresp: &ResponseMsg{
				Message: "Invalid amount",
			},
		},
		{
			reqbody: map[string]interface{}{
				"amount": 70_000,
			},
			mockUpdateAmountAllowedAllowances: &MockSetting{
				Args: []interface{}{
					mock.Anything,
					"k-receipt",
					float64(70_000),
				},
				Returns: []interface{}{
					database.AllowedAllowance{},
					errors.New("an error"),
				},
			},
			want: nil,
			errresp: &ResponseMsg{
				Message: "Failed to update k-receipt amount",
			},
		},
	}

	for i, tc := range tcs {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dbmock := new(AdminDBMock)

			if tc.mockUpdateAmountAllowedAllowances != nil {
				dbmock.On(
					"UpdateAmountAllowedAllowances",
					tc.mockUpdateAmountAllowedAllowances.Args...,
				).Return(tc.mockUpdateAmountAllowedAllowances.Returns...)
			}

			h := NewAdminHandler(validator.New(), dbmock)

			val, _ := json.Marshal(tc.reqbody)

			req := httptest.NewRequest(http.MethodPost, "/admin/deductions/k-receipt", strings.NewReader(string(val)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := echo.New()

			goterr := h.UpdateKReceipt(e.NewContext(req, rec))

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

			var got map[string]float64

			err := json.Unmarshal([]byte(rec.Body.String()), &got)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)

			equal := reflect.DeepEqual(tc.want, got)

			if !equal {
				assert.Fail(t, fmt.Sprintf("expected %v, \nbut got %v", tc.want, got))
			}
		})
	}
}
