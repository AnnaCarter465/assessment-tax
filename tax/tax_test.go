package tax

import (
	"reflect"
	"testing"
)

func TestCalculateTax(t *testing.T) {
	type TC struct {
		name              string
		allowedAllowances Allowances
		income            float64
		allowances        Allowances
		wht               float64
		expectedTax       float64
		expectedRefund    float64
		expectStatements  []TaxStatement
	}

	tcs := []TC{
		{
			name:              "income 500,000 and no allowances", // exp01
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"donation": 0},
			wht:               0,
			expectedTax:       29_000,
			expectedRefund:    0,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  29_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "allowances from user that not in allowed allowances are not calculate",
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"something": 1000},
			wht:               0,
			expectedTax:       29_000,
			expectedRefund:    0,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  29_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "allowances from user that in default allowances are not calculate again",
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"personal": 100_000},
			wht:               0,
			expectedTax:       29_000,
			expectedRefund:    0,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  29_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "income 500,000 and wht 25,000", // exp02
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"donation": 0},
			wht:               25_000,
			expectedTax:       4000,
			expectedRefund:    0,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  29_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "income 500,000 and donation 200,000", // exp03, exp04
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"donation": 200_000},
			wht:               0,
			expectedTax:       19_000,
			expectedRefund:    0,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  19_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "netincome < 0",
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            110_000,
			allowances:        Allowances{"donation": 100_000},
			wht:               20_000,
			expectedTax:       0,
			expectedRefund:    20_000,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "tax < wht",
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"donation": 100_000},
			wht:               60_000,
			expectedTax:       0,
			expectedRefund:    41_000,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  19_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
		},
		{
			name:              "income 500,000 with k-receipt and donation allowances", // exp07
			allowedAllowances: Allowances{"donation": 100_000, "k-receipt": 50_000},
			income:            500_000,
			allowances:        Allowances{"donation": 100_000, "k-receipt": 200_000},
			wht:               0,
			expectedTax:       14_000,
			expectedRefund:    0,
			expectStatements: []TaxStatement{
				{
					Rate: Rate{Percentage: 0, Max: 150_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.1, Max: 500_000},
					Tax:  14_000,
				},
				{
					Rate: Rate{Percentage: 0.15, Max: 1_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.2, Max: 2_000_000},
					Tax:  0,
				},
				{
					Rate: Rate{Percentage: 0.35, Max: -1},
					Tax:  0,
				},
			},
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
			taxer.SetWht(tc.wht)

			for allowanceType, allowanceAmount := range tc.allowances {
				taxer.AddAllowance(allowanceType, allowanceAmount)
			}

			got := taxer.CalculateTaxSummary()

			if got.Tax != tc.expectedTax {
				t.Errorf("Wrong tax expected %v, but got %v", tc.expectedTax, got.Tax)
			}

			if got.Refund != tc.expectedRefund {
				t.Errorf("Wrong refund expected %v, but got %v", tc.expectedRefund, got.Refund)
			}

			if !reflect.DeepEqual(got.TaxStatements, tc.expectStatements) {
				t.Errorf("Wrong tax statements expected %v, but got %v", tc.expectStatements, got.TaxStatements)
			}
		})
	}
}
