// SPDX-License-Identifier: MIT
package plugins

import "testing"

func FuzzValidateRSG(f *testing.F) {
	f.Add([]byte{0x92, 0x01, 0xa3, 0x73, 0x69, 0x67}, []byte("payload"))
	f.Fuzz(func(t *testing.T, rsg, payload []byte) {
		_, _ = ValidateRSG(rsg, payload, nil)
	})
}
