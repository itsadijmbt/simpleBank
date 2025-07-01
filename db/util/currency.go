package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	INR = "INR"
)

func IsSupportedCurrency(curr string) bool {

	switch curr {

	case USD, EUR, CAD, INR:
		return true

	}
	return false
}
