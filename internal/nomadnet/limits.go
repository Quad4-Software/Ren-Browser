// SPDX-License-Identifier: MIT
package nomadnet

import "fmt"

func CheckResponseSize(resp []byte, total int64, maxBytes int) error {
	if maxBytes <= 0 {
		return nil
	}
	if total > int64(maxBytes) {
		return fmt.Errorf("response too large: advertised %d bytes (limit %d)", total, maxBytes)
	}
	if len(resp) > maxBytes {
		return fmt.Errorf("response too large: received %d bytes (limit %d)", len(resp), maxBytes)
	}
	return nil
}
