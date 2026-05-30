// Package season provides constants and utilities for the AniList season
// calendar.
//
// AniList defines seasons by calendar month, using the northern-hemisphere
// meteorological seasons:
//
//	WINTER  — December, January, February   (peak: early January)
//	SPRING  — March, April, May             (peak: early April)
//	SUMMER  — June, July, August            (peak: early July)
//	FALL    — September, October, November  (peak: early October)
//
// The seasonYear field on AniList media is the calendar year the season falls
// under. This means WINTER 2026 covers January–February 2026 (not December
// 2025). December anime get the current year's WINTER season, so a show
// premiering December 2025 would be WINTER 2025, not WINTER 2026.
package season

import "time"

// Season constants as used by the AniList GraphQL API.
const (
	// Winter covers December, January, February.
	// Most anime premieres fall in early January.
	Winter = "WINTER"

	// Spring covers March, April, May.
	// Most anime premieres fall in early April.
	Spring = "SPRING"

	// Summer covers June, July, August.
	// Most anime premieres fall in early July.
	Summer = "SUMMER"

	// Fall covers September, October, November.
	// Most anime premieres fall in early October.
	Fall = "FALL"
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
