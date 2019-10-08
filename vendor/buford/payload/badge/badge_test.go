package badge_test

import (
	"fmt"
	"testing"

	"github.com/RobotsAndPencils/buford/payload/badge"
)

func Example() {
	var b badge.Badge
	fmt.Println(b)

	fmt.Println(badge.Preserve)
	fmt.Println(badge.Clear)
	fmt.Println(badge.New(42))

	// Output:
	// preserve
	// preserve
	// set 0
	// set 42
}

func TestDefaultBadge(t *testing.T) {
	b := badge.Badge{}
	if _, ok := b.Number(); ok {
		t.Errorf("Expected badge number to be omitted.")
	}
}

func TestPreserveBadge(t *testing.T) {
	b := badge.Preserve
	if _, ok := b.Number(); ok {
		t.Errorf("Expected badge number to be omitted.")
	}
}

func TestClearBadge(t *testing.T) {
	b := badge.Clear
	n, ok := b.Number()
	if !ok {
		t.Errorf("Expected badge to be set for removal.")
	}
	if n != 0 {
		t.Errorf("Expected badge number to be 0, got %d.", n)
	}
}

func TestNewBadge(t *testing.T) {
	b := badge.New(4)
	n, ok := b.Number()
	if !ok {
		t.Errorf("Expected badge to be set to change.")
	}
	if n != 4 {
		t.Errorf("Expected badge number to be 4, got %d.", n)
	}
}
