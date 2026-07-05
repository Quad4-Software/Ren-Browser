// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "os"

func regressionMarkupSeedsStatic() []string {
	var out []string
	err := withRegressionRoot(func(root *os.Root) error {
		names, err := listRegressionMuNames(root)
		if err != nil {
			return err
		}
		out = make([]string, 0, len(names))
		for _, name := range names {
			raw, err := readRegressionFile(root, name)
			if err != nil {
				continue
			}
			out = append(out, string(raw))
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return out
}
