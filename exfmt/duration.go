// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exfmt

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var Day = 24 * time.Hour
var Week = 7 * Day

type Pluralizer func(int) string

func Pluralizable(unit string) Pluralizer {
	return func(value int) string {
		if value == 1 {
			return "1 " + unit
		}
		return fmt.Sprintf("%d %ss", value, unit)
	}
}

func NonPluralizable(unit string) Pluralizer {
	return func(value int) string {
		return fmt.Sprintf("%d %s", value, unit)
	}
}

func Duration(d time.Duration) string {
	return DurationCustom(d, nil, Week, Day, time.Hour, time.Minute, time.Second)
}

func appendDurationPart(time, unit time.Duration, name Pluralizer, parts *[]string) (remainder time.Duration) {
	if time < unit {
		return time
	}
	value := int(time / unit)
	remainder = time % unit
	*parts = append(*parts, name(value))
	return
}

var DefaultDurationUnitNames = map[time.Duration]Pluralizer{
	Week:             Pluralizable("week"),
	Day:              Pluralizable("day"),
	time.Hour:        Pluralizable("hour"),
	time.Minute:      Pluralizable("minute"),
	time.Second:      Pluralizable("second"),
	time.Millisecond: NonPluralizable("ms"),
	time.Microsecond: NonPluralizable("µs"),
	time.Nanosecond:  NonPluralizable("ns"),
}

func DurationCustom(d time.Duration, names map[time.Duration]Pluralizer, units ...time.Duration) string {
	if d < 0 {
		panic(errors.New("exfmt.Duration: negative duration"))
	} else if len(units) == 0 {
		panic(errors.New("exfmt.Duration: no units provided"))
	} else if d < units[len(units)-1] {
		return "now"
	}
	if names == nil {
		names = DefaultDurationUnitNames
	}
	parts := make([]string, 0, 2)
	for _, unit := range units {
		name, ok := names[unit]
		if !ok {
			panic(fmt.Errorf("exfmt.Duration: no name for unit %q", unit))
		}
		d = appendDurationPart(d, unit, name, &parts)
	}
	if len(parts) > 2 {
		parts[0] = strings.Join(parts[:len(parts)-1], ", ")
		parts[1] = parts[len(parts)-1]
		parts = parts[:2]
	}
	return strings.Join(parts, " and ")
}
