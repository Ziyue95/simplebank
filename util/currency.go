package util

// Constants for all supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	// use a simple switch case statement
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
