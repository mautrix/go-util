// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package dbutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dialectFilterTest struct {
	name      string
	line      string
	dialect   Dialect
	count     int
	uncomment bool
}

func TestParseDialectFilter(t *testing.T) {
	db := &Database{Dialect: SQLite}
	tests := []dialectFilterTest{
		{"Own dialect: single line", `-- only: sqlite`, SQLite, 1, false},
		{"Own dialect: multiple lines", `-- only: sqlite for next 5 lines`, SQLite, 5, false},
		{"Own dialect: fenced", `-- only: sqlite until "end only"`, SQLite, -1, false},
		{"Own dialect: single line, commented", `-- only: sqlite (line commented)`, SQLite, 1, true},
		{"Own dialect: multiple lines, commented", `-- only: sqlite for next 5 lines (lines commented)`, SQLite, 5, true},
		{"Own dialect: fenced, commented", `-- only: sqlite until "end only" (lines commented)`, SQLite, -1, true},

		{"Other dialect: single line", `-- only: postgres`, Postgres, 1, false},
		{"Other dialect: multiple lines", `-- only: postgres for next 5 lines`, Postgres, 5, false},
		{"Other dialect: fenced", `-- only: postgres until "end only"`, Postgres, -1, false},
		{"Other dialect: single line, commented", `-- only: postgres (line commented)`, Postgres, 1, true},
		{"Other dialect: multiple lines, commented", `-- only: postgres for next 5 lines (lines commented)`, Postgres, 5, true},
		{"Other dialect: fenced, commented", `-- only: postgres until "end only" (lines commented)`, Postgres, -1, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialect, lines, uncomment, err := db.parseDialectFilter([]byte(test.line))
			assert.NoError(t, err)
			assert.Equal(t, test.dialect, dialect)
			assert.Equal(t, test.count, lines)
			assert.Equal(t, test.uncomment, uncomment)
		})
	}
}

type filterOutputTest struct {
	name           string
	input          string
	outputPostgres string
	outputSQLite   string
}

func TestFilterSQLUpgrade(t *testing.T) {
	pg := &Database{Dialect: Postgres}
	lite := &Database{Dialect: SQLite}
	tests := []filterOutputTest{
		{"Single line, commented", `
			-- only: postgres
			meowgres
			-- only: sqlite (lines commented)
--			meowlite
		`, `
			meowgres
		`, `
			meowlite
		`,
		},
		{"Fenced, commented", `
			-- only: postgres until "end only"
			meowgres
			line 2
			-- end only postgres
			shared line
			-- only: sqlite until "end only" (lines commented)
--			meowlite
--			line 2.5
			-- end only sqlite
			shared line 2
		`, `
			meowgres
			line 2
			shared line
			shared line 2
		`, `
			shared line
			meowlite
			line 2.5
			shared line 2
		`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := pg.filterSQLUpgrade(bytes.Split([]byte(test.input), []byte("\n")))
			assert.NoError(t, err)
			assert.Equal(t, test.outputPostgres, out)
			out, err = lite.filterSQLUpgrade(bytes.Split([]byte(test.input), []byte("\n")))
			assert.NoError(t, err)
			assert.Equal(t, test.outputSQLite, out)
		})
	}
}
