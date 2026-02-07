// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"sync"
	"time"
)

var (
	testAccSuffixOnce sync.Once
	testAccSuffix     string
)

func testAccRunSuffix() string {
	testAccSuffixOnce.Do(func() {
		// yymmddhhmmss keeps IDs compact while staying human-readable.
		testAccSuffix = time.Now().UTC().Format("060102150405")
	})

	return testAccSuffix
}

func testAccIdentifier(prefix string, maxLen int) string {
	id := prefix + testAccRunSuffix()
	if len(id) > maxLen {
		return id[:maxLen]
	}

	return id
}
