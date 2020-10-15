// Package badge allows you to preserve, set or clear the number displayed
// on your App icon.
package badge

import "fmt"

// Badge number to display on the App icon.
type Badge struct {
	number uint
	isSet  bool
}

// Preserve the current badge (default behavior).
var Preserve = Badge{}

// Clear the badge.
var Clear = Badge{number: 0, isSet: true}

// New badge with a set value.
// A badge set to 0 is the same as badge.Clear.
func New(number uint) Badge {
	return Badge{number: number, isSet: true}
}

// Number to display on the App Icon and if should be changed.
// If the badge should not be changed, the number has no effect.
func (b *Badge) Number() (uint, bool) {
	return b.number, b.isSet
}

// String prints out a badge
func (b Badge) String() string {
	if b.isSet {
		return fmt.Sprintf("set %d", b.number)
	}
	return "preserve"
}
