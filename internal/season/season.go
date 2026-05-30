package season

import "time"

// Season constants as used by the AniList GraphQL API.
const (
	Winter = "WINTER"
	Spring = "SPRING"
	Summer = "SUMMER"
	Fall   = "FALL"
)

// All returns all four seasons in calendar order.
func All() []string {
	return []string{Winter, Spring, Summer, Fall}
}

// Detect returns the current season based on calendar month.
func Detect() string {
	switch time.Now().Month() {
	case time.December, time.January, time.February:
		return Winter
	case time.March, time.April, time.May:
		return Spring
	case time.June, time.July, time.August:
		return Summer
	default:
		return Fall
	}
}
