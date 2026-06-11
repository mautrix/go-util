package dbutil

import (
	"regexp"
	"testing"
)

var positionalParamRegex = regexp.MustCompile(`\$(\d+)`)

func regexReplacePositionalParams(query string) string {
	return positionalParamRegex.ReplaceAllString(query, "?$1")
}

func TestReplacePositionalParams(t *testing.T) {
	cases := []string{
		"",
		"$",
		"$1",
		"$12",
		"a$",
		"$a",
		"$$1",
		"$1$2$3",
		"$ 1",
		"no placeholders at all",
		"trailing dollar $",
		"SELECT session FROM whatsmeow_sessions WHERE our_jid=$1 AND their_id=$2",
		"INSERT INTO whatsmeow_sessions (our_jid, their_id, session) VALUES ($1, $2, $3) ON CONFLICT (our_jid, their_id) DO UPDATE SET session=excluded.session",
		"SELECT their_id, session FROM whatsmeow_sessions WHERE our_jid=$1 AND their_id IN ($2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)",
		"WHERE a=$2 AND b=$1",
		"cost: US$50 AND id=$7",
		"weird $0 but valid",
		"unicode é$1ç $2ü",
		"$999999999999",
	}
	for _, query := range cases {
		expected := regexReplacePositionalParams(query)
		actual := replacePositionalParams(query)
		if actual != expected {
			t.Errorf("mismatch for %q:\nregexp:  %q\nscanner: %q", query, expected, actual)
		}
	}
}

// Generated coverage: every 3-character combination of a small relevant
// alphabet must behave identically to the regexp it replaces.
func TestReplacePositionalParamsExhaustive(t *testing.T) {
	alphabet := []byte{'$', '1', '0', 'a', ' ', '?'}
	var build func(prefix string, depth int)
	build = func(prefix string, depth int) {
		expected := regexReplacePositionalParams(prefix)
		actual := replacePositionalParams(prefix)
		if actual != expected {
			t.Errorf("mismatch for %q: regexp=%q scanner=%q", prefix, expected, actual)
		}
		if depth == 0 {
			return
		}
		for _, c := range alphabet {
			build(prefix+string(c), depth-1)
		}
	}
	build("", 4)
}

func TestReplacePositionalParamsNoAllocWithoutPlaceholders(t *testing.T) {
	query := "SELECT * FROM table WHERE nothing"
	allocs := testing.AllocsPerRun(100, func() {
		_ = replacePositionalParams(query)
	})
	if allocs != 0 {
		t.Errorf("expected 0 allocations for query without placeholders, got %.1f", allocs)
	}
}

func BenchmarkReplacePositionalParams(b *testing.B) {
	query := "INSERT INTO whatsmeow_sessions (our_jid, their_id, session) VALUES ($1, $2, $3) ON CONFLICT (our_jid, their_id) DO UPDATE SET session=excluded.session"
	b.Run("scanner", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = replacePositionalParams(query)
		}
	})
	b.Run("regexp", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = regexReplacePositionalParams(query)
		}
	})
}
