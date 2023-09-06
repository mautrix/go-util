// Copyright (c) 2021 Dillon Dixon
// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package retryafter

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackoffFromResponse(t *testing.T) {
	currentTime := time.Now().Truncate(time.Second)
	now = func() time.Time {
		return currentTime
	}

	defaultBackoff := time.Duration(123)

	for name, tt := range map[string]struct {
		headerValue string
		expected    time.Duration
	}{
		"AsDate": {
			headerValue: currentTime.In(time.UTC).Add(5 * time.Hour).Format(http.TimeFormat),
			expected:    time.Duration(5) * time.Hour,
		},
		"AsSeconds": {
			headerValue: "12345",
			expected:    time.Duration(12345) * time.Second,
		},
		"Missing": {
			headerValue: "",
			expected:    defaultBackoff,
		},
		"Bad": {
			headerValue: "invalid",
			expected:    defaultBackoff,
		},
	} {
		t.Run(name, func(t *testing.T) {
			parsed := Parse(tt.headerValue, defaultBackoff)
			assert.Equal(t, tt.expected, parsed)
		})
	}
}
