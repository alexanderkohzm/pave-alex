package models

import (
	"testing"
)

func TestCurrency_IsValid(t *testing.T) {
	tests := []struct {
		name          string
		currency      Currency
		expectedValue bool
	}{
		{"USD", USD, true},
		{"GEL", GEL, true},
		{"SGD", "SGD", false},
		{"", "emptystring", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.currency.IsValid(); got != test.expectedValue {
				t.Errorf("IsValid() = %v, expected %v", got, test.expectedValue)
			}
		})
	}
}

func TestNewMoney(t *testing.T) {
	tests := []struct {
		testName       string
		amount         float64
		currency       Currency
		expectedAmount int64
		expectedError  bool
	}{
		{
			testName:       "valid USD amount",
			amount:         10.50,
			currency:       USD,
			expectedAmount: 1050, // 10.50 USD = 1050 cents
			expectedError:  false,
		},
		{
			testName:       "valid GEL amount",
			amount:         15.75,
			currency:       GEL,
			expectedAmount: 1575, // 15.75 GEL = 1575 tetri
			expectedError:  false,
		},
		{
			testName:       "zero amount",
			amount:         0,
			currency:       USD,
			expectedAmount: 0,
			expectedError:  false,
		},
		{
			testName:       "one cent",
			amount:         0.01,
			currency:       USD,
			expectedAmount: 1,
			expectedError:  false,
		},
		{
			testName:       "less than one cent",
			amount:         0.001,
			currency:       USD,
			expectedAmount: 0,
			expectedError:  false,
		},
		{
			testName:       "invalid currency",
			amount:         10.50,
			currency:       Currency("EUR"),
			expectedAmount: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := NewMoney(tt.amount, tt.currency)
			if (err != nil) != tt.expectedError {
				t.Errorf("NewMoney() error = %v, wantErr %v", err, tt.expectedAmount)
				return
			}
			if !tt.expectedError && got.Amount != tt.expectedAmount {
				t.Errorf("NewMoney() = %v, want %v", got.Amount, tt.expectedAmount)
			}
		})
	}
}

func TestMoney_Convert(t *testing.T) {

	tests := []struct {
		testName       string
		amount         Money
		rate           float64
		targetCurrency Currency
		expectedValue  int64
		expectedError  bool
	}{
		{
			testName:       "convert USD to GEL",
			amount:         Money{Amount: 100, Currency: USD},
			rate:           2.76,
			targetCurrency: GEL,
			expectedValue:  276,
			expectedError:  false,
		},
		{
			testName:       "convert USD to GEL",
			amount:         Money{Amount: 2760, Currency: USD},
			rate:           1 / 2.76,
			targetCurrency: GEL,
			expectedValue:  1000,
			expectedError:  false,
		},
		{
			testName:       "negative exchange rate",
			amount:         Money{Amount: 1000, Currency: USD},
			rate:           -1.0,
			targetCurrency: GEL,
			expectedValue:  0,
			expectedError:  true,
		},
		{
			testName:       "invalid target currency",
			amount:         Money{Amount: 1000, Currency: GEL},
			rate:           0.5,
			targetCurrency: "SGD",
			expectedValue:  500,
			expectedError:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := tt.amount.Convert(tt.rate, tt.targetCurrency)
			if (err != nil) != tt.expectedError {
				t.Errorf("Convert() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if !tt.expectedError && got.Amount != tt.expectedValue {
				t.Errorf("Convert() = %v, want %v", got.Amount, tt.expectedValue)
			}
		})
	}
}

func TestInt64Pow(t *testing.T) {
	tests := []struct {
		name          string
		a             int32
		b             int32
		expectedValue int64
	}{
		{
			name:          "10^2",
			a:             10,
			b:             2,
			expectedValue: 100,
		},
		{
			name:          "10^0",
			a:             10,
			b:             0,
			expectedValue: 1,
		},
		{
			name:          "2^3",
			a:             2,
			b:             3,
			expectedValue: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int64Pow(tt.a, tt.b); got != tt.expectedValue {
				t.Errorf("int64Pow() = %v, expectedValue %v", got, tt.expectedValue)
			}
		})
	}
}
