package tax

type Rate struct {
	Percentage float64
	Max        float64
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

func (t *Tax) CalculateTax() float64 {
	netIncome := t.income - t.calculateTotalAllowance()

	if netIncome <= 0 {
		return 0
	}

	var tax float64

	remain := netIncome

	for _, rate := range t.taxConf.Rates {

		// highest stage
		if netIncome <= rate.Max {
			tax += remain * rate.Percentage
			remain = 0
			break
		}

		// infinite stage
		if rate.Max == -1 {
			tax += remain * rate.Percentage
			remain = 0
			break
		}

		tax += rate.Max * rate.Percentage
		remain -= rate.Max
	}

	return tax
}
