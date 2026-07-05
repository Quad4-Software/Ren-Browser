// SPDX-License-Identifier: MIT
package micron_test

import (
	"testing"

	"renbrowser/internal/micron"
)

func FuzzResolveNavigation(f *testing.F) {
	f.Add("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu", ":page/submit.mu`action=run", "user", "alice")
	f.Add("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu", "/page/other.mu", "", "")

	f.Fuzz(func(t *testing.T, current, destination, fieldName, fieldValue string) {
		var inputs []micron.FieldInput
		if fieldName != "" {
			inputs = []micron.FieldInput{{Type: "text", Name: fieldName, Value: fieldValue}}
		}
		next, err := micron.ResolveNavigation(current, destination, fieldName, inputs)
		if err != nil {
			return
		}
		if next == "" {
			t.Fatal("empty navigation url")
		}
	})
}
