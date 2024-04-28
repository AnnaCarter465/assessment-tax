package tax

type Rate struct {
	Percentage float64
	Max        float64
	Label      string
}
type Allowances map[string]float64

type TaxConfig struct {
	Rates             []Rate
	AllowedAllowances Allowances // allowed allowances with maximum amount
	DefaultAllowances Allowances
}

type Tax struct {
	income     float64
	allowances Allowances
	taxConf    TaxConfig
	wht        float64
}

func NewTax(taxConf TaxConfig) *Tax {
	return &Tax{
		allowances: make(Allowances),
		taxConf:    taxConf,
	}
}

func (t *Tax) SetIncome(income float64) *Tax {
	t.income = income
	return t
}

func (t *Tax) SetWht(wht float64) *Tax {
	t.wht = wht
	return t
}

func (t *Tax) AddAllowance(allowanceType string, amount float64) *Tax {
	t.allowances[allowanceType] = amount
	return t
}

func (t *Tax) calculateTotalAllowance() float64 {
	var totalAllowance float64

	for _, allowanceAmount := range t.taxConf.DefaultAllowances {
		totalAllowance += allowanceAmount
	}

	for allowanceType, allowanceAmount := range t.allowances {
		// check if allowances input is duplicated with default allowance, we should ignore it.
		_, ok := t.taxConf.DefaultAllowances[allowanceType]

		if ok {
			continue
		}

		// check if provided allowances are allowed and they shouldn't go over max amount
		maxAmount, ok := t.taxConf.AllowedAllowances[allowanceType]

		if !ok {
			continue
		}

		amount := allowanceAmount

		if amount > maxAmount {
			amount = maxAmount
		}

		totalAllowance += amount
	}

	return totalAllowance
}

type TaxStatement struct {
	Rate Rate
	Tax  float64
}

func (t *Tax) calculateTaxStatement(netIncome float64) []TaxStatement {
	var ts []TaxStatement

	var totalTax float64

	remain := netIncome

	for _, rate := range t.taxConf.Rates {

		if remain <= 0 {
			ts = append(ts, TaxStatement{
				Rate: rate,
				Tax:  0,
			})

			continue
		}

		// highest stage or infinity stage
		if netIncome <= rate.Max || rate.Max == -1 {
			tax := remain * rate.Percentage
			totalTax += tax
			remain = 0

			ts = append(ts, TaxStatement{
				Rate: rate,
				Tax:  tax,
			})

			continue
		}

		tax := rate.Max * rate.Percentage

		totalTax += tax
		remain -= rate.Max

		ts = append(ts, TaxStatement{
			Rate: rate,
			Tax:  tax,
		})
	}

	return ts
}

type TaxSummary struct {
	TaxStatements []TaxStatement
	Tax           float64
	Refund        float64
}

func (t *Tax) CalculateTaxSummary() TaxSummary {
	netIncome := t.income - t.calculateTotalAllowance()

	if netIncome <= 0 {
		return TaxSummary{
			TaxStatements: nil,
			Tax:           0,
			Refund:        t.wht,
		}
	}

	statements := t.calculateTaxStatement(netIncome)

	var tax float64

	for _, statement := range statements {
		tax += statement.Tax
	}

	var refund float64
	if tax <= t.wht {
		refund = t.wht - tax
		tax = 0
	} else {
		tax = tax - t.wht
	}

	return TaxSummary{
		TaxStatements: statements,
		Tax:           tax,
		Refund:        refund,
	}
}
