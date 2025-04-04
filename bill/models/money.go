package models

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

type Currency string

const (
	USD Currency = "USD"
	GEL Currency = "GEL"
)

var validCurrencies = []Currency{USD, GEL}

func (c Currency) Validate() bool {
	// validation - can only be either USD or GEL
	// for loop
	for _, currency := range validCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

func (c Currency) IsValid() bool {
	_, exists := DecimalPlaces[string(c)]
	return exists
}

var DecimalPlaces = map[string]int{
	"USD": 2, // cents: 100 cents = 1 USD
	"GEL": 2, // tetri: 100 tetri = 1 GEL
}

type Money struct {
	Amount   int64    `json:"amount"`   // stored in lowest unit of currency
	Currency Currency `json:"currency"` // e.g. "USD"
}

var ExchangeRates = map[string]map[string]float64{
	"USD": {
		"GEL": 2.76,
	},
	"GEL": {
		"USD": 1 / 2.76,
	},
}

func NewMoney(amount float64, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, fmt.Errorf("unsupported currency: %s", currency)
	}

	decimals := DecimalPlaces[string(currency)]
	smallestUnit := amount * math.Pow10(decimals)

	return Money{
		Amount:   int64(smallestUnit),
		Currency: currency,
	}, nil
}

func int64Pow(a, b int32) int64 {
	result := int64(1)
	for i := int32(0); i < b; i++ {
		result *= int64(a)
	}
	return result
}

func (m *Money) Convert(rate float64, targetCurrency Currency) (*Money, error) {
	if rate <= 0 {
		return nil, fmt.Errorf("exchange rate must be positive")
	}

	// if target currency is the same as the source currency, return the original money
	if m.Currency == targetCurrency {
		return m, nil
	}

	sourceDecimalPlaces, ok := DecimalPlaces[string(m.Currency)]
	if !ok {
		return nil, fmt.Errorf("invalid source currency: %s", m.Currency)
	}
	targetDecimalPlaces, ok := DecimalPlaces[string(targetCurrency)]
	if !ok {
		return nil, fmt.Errorf("invalid target currency: %s", targetCurrency)
	}

	// Convert Amount (int64) to decimal in units
	units := decimal.NewFromInt(m.Amount).Div(decimal.NewFromInt(int64Pow(10, int32(sourceDecimalPlaces))))
	// Apply exchange rate
	// need to convert int to decimal to perform multiplication
	// Shopspring requires int32
	rateDecimal := decimal.NewFromFloat(rate)
	convertedUnits := units.Mul(rateDecimal).Round(int32(targetDecimalPlaces))
	// Convert back to smallest unit for target currency
	multiplier := decimal.NewFromInt(int64Pow(10, int32(targetDecimalPlaces)))
	convertedAmount := convertedUnits.Mul(multiplier).Round(0).IntPart()

	return &Money{
		Amount:   convertedAmount,
		Currency: targetCurrency,
	}, nil
}
