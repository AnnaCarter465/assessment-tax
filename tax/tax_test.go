package tax

import (
	"testing"
)

func TestCalculateTax2(t *testing.T) {
	type TC struct {
		name              string
		allowedAllowances Allowances
		income            float64
		allowances        Allowances
		wht               float64
		expectedTax       float64
	}

	tcs := []TC{
		{
			name:              "income 500,000 and no allowances", // exp01
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"donation": 0},
			wht:               0,
			expectedTax:       29_000,
		},
		{
			name:              "allowances from user that not in allowed allowances are not calculate",
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"something": 1000},
			wht:               0,
			expectedTax:       29_000,
		},
		{
			name:              "allowances from user that in default allowances are not calculate again",
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"personal": 100_000},
			wht:               0,
			expectedTax:       29_000,
		},
	}

	t.Parallel()

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			taxer := NewTax(
				TaxConfig{
					Rates: []Rate{
						{Percentage: 0, Max: 150_000},
						{Percentage: 0.1, Max: 500_000},
						{Percentage: 0.15, Max: 1_000_000},
						{Percentage: 0.2, Max: 2_000_000},
						{Percentage: 0.35, Max: -1},
					},
					DefaultAllowances: Allowances{"personal": 60_000},
					AllowedAllowances: tc.allowedAllowances,
				},
			)

			taxer.SetIncome(tc.income)

			for allowanceType, allowanceAmount := range tc.allowances {
				taxer.AddAllowance(allowanceType, allowanceAmount)
			}

			got := taxer.CalculateTax()

			if got != tc.expectedTax {
				t.Errorf("Wrong result expected %v, but got %v", tc.expectedTax, got)
			}
		})
	}
}
