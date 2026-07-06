// SPDX-License-Identifier: MIT
package content

import _ "embed"

//go:embed editor_template.mu
var defaultEditorTemplate string

func DefaultEditorTemplate() string {
	return defaultEditorTemplate
}
