package season

import (
	"fmt"
	"strings"
	"time"
)

// Season enum values as used by the AniList GraphQL API.
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

// Resolve returns the list of seasons to process.
// If seasonArg is empty, returns all four. Otherwise validates and returns
// the single season (case-insensitive).
func Resolve(year int, seasonArg string) []string {
	if seasonArg == "" {
		return All()
	}

	switch strings.ToLower(seasonArg) {
	case "winter":
		return []string{Winter}
	case "spring":
		return []string{Spring}
	case "summer":
		return []string{Summer}
	case "fall":
		return []string{Fall}
	default:
		// Will be caught by caller as empty slice
		return nil
	}
}

// ValidateYear returns an error if the year is unrealistic (before 2000 or after 2100).
func ValidateYear(year int) error {
	if year < 2000 || year > 2100 {
		return fmt.Errorf("year %d is out of range (2000-2100)", year)
	}
	return nil
}
