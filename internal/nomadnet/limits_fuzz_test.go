// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"strings"
	"testing"

	"renbrowser/internal/nomadnet"
)

func FuzzCheckResponseSize(f *testing.F) {
	f.Add([]byte("hello"), int64(5), 8)
	f.Add([]byte(""), int64(0), 0)
	f.Fuzz(func(t *testing.T, data []byte, total int64, max int) {
		if max < 0 {
			max = -max
		}
		if total < 0 {
			total = -total
		}
		err := nomadnet.CheckResponseSize(data, total, max)
		if max <= 0 {
			if err != nil {
				t.Fatalf("max=%d err=%v", max, err)
			}
			return
		}
		if int64(len(data)) > int64(max) || total > int64(max) {
			if err == nil {
				t.Fatalf("expected error for len=%d total=%d max=%d", len(data), total, max)
			}
			if !strings.Contains(err.Error(), "response too large") {
				t.Fatalf("err = %v", err)
			}
			return
		}
		if err != nil {
			t.Fatalf("unexpected err=%v len=%d total=%d max=%d", err, len(data), total, max)
		}
	})
}
