// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package emojishortcodes

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"go.mau.fi/util/exstrings"
	"go.mau.fi/util/variationselector"
)

//go:embed shortcodes.json
//go:generate go run shortcode_gen.go
var shortcodesJSON string
var shortcodeMap = map[string]string{}
var shortcodeInit = sync.OnceFunc(func() {
	if err := json.Unmarshal(exstrings.UnsafeBytes(shortcodesJSON), &shortcodeMap); err != nil {
		panic(fmt.Errorf("failed to unmarshal shortcodes: %w", err))
	}
})

func GetMap() map[string]string {
	shortcodeInit()
	return shortcodeMap
}

var skinToneRemover = strings.NewReplacer(
	"\U0001F3FB", "",
	"\U0001F3FC", "",
	"\U0001F3FD", "",
	"\U0001F3FE", "",
	"\U0001F3FF", "",
)

func Get(emoji string) string {
	sc := GetMap()
	emoji = skinToneRemover.Replace(emoji)
	return sc[variationselector.Add(emoji)]
}
