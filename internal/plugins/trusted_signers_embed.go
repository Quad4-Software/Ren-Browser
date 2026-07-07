// SPDX-License-Identifier: MIT
package plugins

import _ "embed"

//go:embed trusted_signers.json
var trustedSignersJSON []byte
