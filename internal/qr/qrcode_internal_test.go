// SPDX-License-Identifier: MIT

package qr

import "testing"

func TestFinderPatterns(t *testing.T) {
	m, err := encodeMatrix("renbrowser")
	if err != nil {
		t.Fatal(err)
	}
	checkFinder := func(x, y int) {
		t.Helper()
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 7; dx++ {
				want := dx == 0 || dx == 6 || dy == 0 || dy == 6 || (dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4)
				if m.isDark(x+dx, y+dy) != want {
					t.Fatalf("finder at (%d,%d) module (%d,%d): got %v want %v", x, y, dx, dy, !want, want)
				}
			}
		}
	}
	checkFinder(0, 0)
	checkFinder(m.size-7, 0)
	checkFinder(0, m.size-7)
}
