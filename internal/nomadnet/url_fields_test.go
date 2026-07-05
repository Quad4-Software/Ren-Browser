// SPDX-License-Identifier: MIT
package nomadnet

import "testing"

func TestFormatURLWithFields(t *testing.T) {
	got := FormatURLWithFields("abc", "/page/x.mu", map[string]string{"a": "1", "b": "2"})
	want := "abc:/page/x.mu`a=1|b=2"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
