// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package redact_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"go.mau.fi/util/redact"
)

// relaxed is the kind of policy a caller builds for structure-heavy payloads
// like Bloks trees, where labels and identifiers need to stay readable.
func relaxed() redact.Policy {
	p := redact.MakeDefaultPolicy()
	p.KeepProse = true
	p.MaxDigits = 10
	return p
}

func TestString_MarkerFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *regexp.Regexp
	}{
		{"opaque", "aXk39fJs01Zq7wLmPpQ2", regexp.MustCompile(`^redacted_20char_str_[0-9a-f]{8}$`)},
		{"hex", "deadbeefcafef00dba5eba11", regexp.MustCompile(`^redacted_24char_hex_[0-9a-f]{8}$`)},
		{"uuid", "a1b2c3d4-e5f6-1234-9abc-def012345678", regexp.MustCompile(`^redacted_36char_uuid_lowercase_[0-9a-f]{8}$`)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := redact.MakeDefaultPolicy().String(tc.input)
			if !tc.want.MatchString(got) {
				t.Fatalf("String(%q) = %q, want match %s", tc.input, got, tc.want)
			}
		})
	}
}

func TestString_ProseKeepsLabelsNotTokens(t *testing.T) {
	p := relaxed()
	for _, label := range []string{"true", "100%", "N/A", "#FF0000", "Log in", "en_US"} {
		if got := p.String(label); got != label {
			t.Errorf("short label %q should be kept, got %q", label, got)
		}
	}
	for _, token := range []string{"aXk39fJs01Zq", "AbCdEf123", "Zq7wLmP", "$#!@%^&*("} {
		if got := p.String(token); !isRedacted(got) {
			t.Errorf("short opaque token %q should be hashed, got %q", token, got)
		}
	}
}

func TestString_Prose(t *testing.T) {
	const sentence = "This is a human readable error message."
	if got := relaxed().String(sentence); got != sentence {
		t.Errorf("relaxed kept prose wrong: got %q", got)
	}
	if got := redact.MakeDefaultPolicy().String(sentence); got == sentence {
		t.Errorf("default should redact prose, got it verbatim: %q", got)
	}
}

func TestString_KeepPatterns(t *testing.T) {
	const internal = "com.bloks.www.aaaaaaaaaaaaaaaaaaaaaaaa"
	p := redact.MakeDefaultPolicy()
	p.KeepPatterns = []*regexp.Regexp{regexp.MustCompile(`^com\.bloks\.`)}
	if got := p.String(internal); got != internal {
		t.Errorf("KeepPatterns should keep %q, got %q", internal, got)
	}
	if got := redact.MakeDefaultPolicy().String(internal); got == internal {
		t.Errorf("without pattern default should redact %q, got it verbatim", internal)
	}
}

func TestString_URL(t *testing.T) {
	const cdnURL = "https://scontent.cdninstagram.com/v/t51/some_opaque_path_1234567890abcdefghij"
	got := relaxed().String(cdnURL)
	const wantPrefix = "https://scontent.cdninstagram.com/"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("URL host should be kept, got %q", got)
	}
	if got == cdnURL {
		t.Errorf("URL path should be redacted, got it verbatim")
	}

	p := relaxed()
	p.KeepURLPathPatterns = []*regexp.Regexp{regexp.MustCompile(`^rsrc\.php/`)}
	const staticURL = "https://static.xx.fbcdn.net/rsrc.php/v3iABC/yk/l/en_US/somebundle.js"
	if got := p.String(staticURL); got != staticURL {
		t.Errorf("allowlisted URL path should be kept whole, got %q", got)
	}
}

func TestString_HashPatternsInsideKeptProse(t *testing.T) {
	tests := []struct{ name, input, pii string }{
		{"email", "We sent a code to john.doe@example.com", "john.doe@example.com"},
		{"phone", "Call us on +1 555 123 4567 today", "+1 555 123 4567"},
		{"otp", "Your code is 123456", "123456"},
		{"card", "Card ending 4111 1111 1111 1111", "4111 1111 1111 1111"},
		{"ipv4", "Signed in from 192.168.1.20", "192.168.1.20"},
	}
	p := relaxed()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := p.String(tc.input)
			if strings.Contains(got, tc.pii) {
				t.Fatalf("%q leaked in %q", tc.pii, got)
			}
			want := strings.Replace(tc.input, tc.pii, redact.MakeDefaultPolicy().String(tc.pii), 1)
			if got != want {
				t.Errorf("String(%q) = %q, want %q", tc.input, got, want)
			}
		})
	}
}

func TestString_HashPatternsEntireString(t *testing.T) {
	p := relaxed()
	for _, pii := range []string{"+15551234567", "jo@ex.com", "123456"} {
		if got, want := p.String(pii), redact.MakeDefaultPolicy().String(pii); got != want {
			t.Errorf("PII %q should become its marker %q, got %q", pii, want, got)
		}
	}
}

func TestString_HashPatternsMergeOverlaps(t *testing.T) {
	p := relaxed()
	got := p.String("Dial +1 555 123 4567 or 123456 now")
	if n := strings.Count(got, "redacted_"); n != 2 {
		t.Errorf("expected two markers, got %d in %q", n, got)
	}
	got = p.String("Ref 123456789")
	if want := "Ref " + redact.MakeDefaultPolicy().String("123456789"); got != want {
		t.Errorf("overlapping matches should collapse into one marker: got %q, want %q", got, want)
	}
}

func TestString_HashPatternsSeeOriginalText(t *testing.T) {
	p := relaxed()
	p.HashPatterns = []*regexp.Regexp{regexp.MustCompile(`\d{6,}`), regexp.MustCompile(`char_`)}
	got := p.String("code 123456")
	if want := "code " + redact.MakeDefaultPolicy().String("123456"); got != want {
		t.Errorf("later patterns must not match inside earlier markers: got %q, want %q", got, want)
	}
}

func TestString_HashPatternsApplyToAllowlists(t *testing.T) {
	p := redact.MakeDefaultPolicy()
	p.KeepPatterns = []*regexp.Regexp{regexp.MustCompile(`^INTERNAL_`)}
	got := p.String("INTERNAL_contact john@example.com")
	if !strings.HasPrefix(got, "INTERNAL_contact ") || strings.Contains(got, "john@example.com") {
		t.Errorf("KeepPatterns string should be kept with PII masked, got %q", got)
	}

	p = redact.MakeDefaultPolicy()
	p.KeepURLPathPatterns = []*regexp.Regexp{regexp.MustCompile(`^rsrc\.php/`)}
	got = p.String("https://static.xx.fbcdn.net/rsrc.php/mail/john@example.com")
	if !strings.HasPrefix(got, "https://static.xx.fbcdn.net/rsrc.php/mail/") || strings.Contains(got, "john@example.com") {
		t.Errorf("allowlisted URL path should be kept with PII masked, got %q", got)
	}
}

func TestString_HashPatternsLeaveWholeHashAlone(t *testing.T) {
	const s = "Call +1 555 123 4567 now aXk39fJs01Zq7wLmPpQ2"
	noPatterns := redact.MakeDefaultPolicy()
	noPatterns.HashPatterns = nil
	if got, want := redact.MakeDefaultPolicy().String(s), noPatterns.String(s); got != want {
		t.Errorf("whole-string marker should not depend on HashPatterns: %q vs %q", got, want)
	}
}

func TestString_HashPatternsNestedJSON(t *testing.T) {
	got := relaxed().String(`{"label":"Sent to john@example.com","count":3}`)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("result should still be JSON: %v (%q)", err, got)
	}
	label, _ := parsed["label"].(string)
	if !strings.HasPrefix(label, "Sent to ") || strings.Contains(label, "john@example.com") {
		t.Errorf("embedded JSON leaf should be kept with PII masked, got %q", label)
	}
}

func TestValue_NumbersRelaxedKeepsTimestamps(t *testing.T) {
	p := relaxed()
	for _, n := range []float64{1726300000, 404, 0} {
		v := any(n)
		p.Value(&v)
		if v != any(n) {
			t.Errorf("relaxed should keep %v, got %v", n, v)
		}
	}
	v := any(float64(123456789012))
	p.Value(&v)
	if got, _ := v.(string); !isRedacted(got) {
		t.Errorf("relaxed should hash a number longer than ten digits, got %v", v)
	}
}

func TestValue_NumbersDefault(t *testing.T) {
	p := redact.MakeDefaultPolicy()
	for _, n := range []any{float64(404), float64(0), float64(4.5), json.Number("-0.75"), int64(9999)} {
		v := n
		p.Value(&v)
		if v != n {
			t.Errorf("%v should be kept, got %v", n, v)
		}
	}
	for _, n := range []any{float64(12345678), float64(51.507351), json.Number("-0.127758"), json.Number("1e21"), int64(10000)} {
		v := n
		p.Value(&v)
		if got, _ := v.(string); !isRedacted(got) {
			t.Errorf("%v should be hashed, got %v", n, v)
		}
	}
	long := any(float64(12345678))
	p.Value(&long)
	if long != any(p.String("12345678")) {
		t.Errorf("hashed number should match its string form, got %v", long)
	}
}

func TestValue_SensitiveKeys(t *testing.T) {
	raw := []byte(`{"email":"john@example.com","count":5,"nested":{"access_token":"tokenvalue123"},"note":"a normal message here"}`)
	out, err := redact.MakeDefaultPolicy().JSON(raw)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if email, _ := got["email"].(string); !isRedacted(email) {
		t.Errorf("email should be redacted, got %q", email)
	}
	if got["count"] != float64(5) {
		t.Errorf("count should be kept, got %v", got["count"])
	}
	nested, _ := got["nested"].(map[string]any)
	if tok, _ := nested["access_token"].(string); !isRedacted(tok) {
		t.Errorf("nested access_token should be redacted, got %q", tok)
	}
}

func TestValue_NameKeysHashedUnderProse(t *testing.T) {
	out, err := relaxed().JSON([]byte(`{"username":"nick_mb","first_name":"Nick","label":"Nick"}`))
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"username", "first_name"} {
		if !isRedacted(got[key]) {
			t.Errorf("%s should be hashed even when prose would keep it, got %q", key, got[key])
		}
	}
	if got["label"] != "Nick" {
		t.Errorf("label is not a name key and should be kept, got %q", got["label"])
	}
}

func TestValue_SensitiveKeyRedactsAnyType(t *testing.T) {
	raw := []byte(`{"token":1234567890,"cookie":["a","b"],"pin":true}`)
	out, err := redact.MakeDefaultPolicy().JSON(raw)
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if tok, _ := got["token"].(string); !isRedacted(tok) {
		t.Errorf("numeric token should be redacted to a marker, got %v", got["token"])
	}
	if ck, _ := got["cookie"].(string); !isRedacted(ck) {
		t.Errorf("array cookie should be redacted to a marker, got %v", got["cookie"])
	}
	if got["pin"] != true {
		t.Errorf("bool under sensitive key is not sensitive on its own, got %v", got["pin"])
	}
}

func TestString_NestedJSON(t *testing.T) {
	nested := `{"ref":"AXk39fJs01Zq7wLmPpQ2rT","label":"Continue"}`
	got := relaxed().String(nested)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("result should still be JSON: %v (%q)", err, got)
	}
	if ref, _ := parsed["ref"].(string); !isRedacted(ref) {
		t.Errorf("opaque value inside embedded JSON should be redacted, got %q", ref)
	}
	if parsed["label"] != "Continue" {
		t.Errorf("short label inside embedded JSON should be kept, got %v", parsed["label"])
	}
}

// A bare numeric string is treated as an opaque string and hashed, not coerced
// into a number, so a long numeric ID does not leak its leading digits.
func TestString_NumericStringHashed(t *testing.T) {
	got := relaxed().String("998877665544332211")
	if !isRedacted(got) {
		t.Fatalf("numeric string should be hashed, got %q", got)
	}
	if !strings.Contains(got, "number") {
		t.Errorf("marker should record the numeric shape, got %q", got)
	}
}

func TestString_Hooks(t *testing.T) {
	keepAll := redact.MakeDefaultPolicy()
	keepAll.ClassifyString = func(string) redact.Decision { return redact.Keep }
	if got := keepAll.String("aXk39fJs01Zq7wLmPpQ2"); got != "aXk39fJs01Zq7wLmPpQ2" {
		t.Errorf("ClassifyString Keep should keep value, got %q", got)
	}
	if got := keepAll.String("Sent to john@example.com"); strings.Contains(got, "john@example.com") {
		t.Errorf("ClassifyString Keep should still mask HashPatterns, got %q", got)
	}

	keepKey := redact.MakeDefaultPolicy()
	keepKey.ClassifyKey = func(key string) redact.Decision {
		if key == "email" {
			return redact.Keep
		}
		return redact.Auto
	}
	out, err := keepKey.JSON([]byte(`{"email":"john@example.com"}`))
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if strings.Contains(string(out), "john@example.com") {
		t.Errorf("ClassifyKey Keep should still mask HashPatterns, got %s", out)
	}

	keepKey.HashPatterns = nil
	out, err = keepKey.JSON([]byte(`{"email":"john@example.com"}`))
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if string(out) != `{"email":"john@example.com"}` {
		t.Errorf("ClassifyKey Keep without HashPatterns should keep sensitive key verbatim, got %s", out)
	}
}

func TestString_Idempotent(t *testing.T) {
	once := redact.MakeDefaultPolicy().String("aXk39fJs01Zq7wLmPpQ2")
	twice := redact.MakeDefaultPolicy().String(once)
	if once != twice {
		t.Errorf("redacting a marker should be a no-op: %q -> %q", once, twice)
	}
}

func TestString_StableAndDistinct(t *testing.T) {
	a1 := redact.MakeDefaultPolicy().String("aXk39fJs01Zq7wLmPpQ2")
	a2 := redact.MakeDefaultPolicy().String("aXk39fJs01Zq7wLmPpQ2")
	if a1 != a2 {
		t.Errorf("same input should hash the same: %q vs %q", a1, a2)
	}
	b := redact.MakeDefaultPolicy().String("bYl48gKt12Ar8xMnQqR3")
	if a1 == b {
		t.Errorf("different inputs should not collide: %q", a1)
	}
}

func TestZeroPolicy_FailsClosed(t *testing.T) {
	var p redact.Policy
	if got := p.String("hello world friend"); !isRedacted(got) {
		t.Errorf("zero policy should redact strings, got %q", got)
	}
	v := any(float64(12345))
	p.Value(&v)
	if got, _ := v.(string); !isRedacted(got) {
		t.Errorf("zero policy should hash numbers, got %v", v)
	}
}

const loginHTML = `<html><head><title>Login</title></head><body>` +
	`<form action="/api/v1/web/accounts/login/ajax/">` +
	`<input name="username" value="secretuser12345678">` +
	`<div class="error">Please enter a valid password.</div>` +
	`<p>Enter the code we sent to +1 555 123 4567.</p>` +
	`</form>` +
	`<script type="application/json">{"config":{"csrf_token":"AbCdEf1234567890XyZ"}}</script>` +
	`</html>`

func TestHTML_Relaxed(t *testing.T) {
	out, err := relaxed().HTML([]byte(loginHTML))
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	got := string(out)
	for _, want := range []string{"<form", "<input", "<title>", "Login", "password", "Enter the code we sent to "} {
		if !strings.Contains(got, want) {
			t.Errorf("structure/prose should be preserved, missing %q in:\n%s", want, got)
		}
	}
	for _, leak := range []string{"secretuser12345678", "AbCdEf1234567890XyZ", "+1 555 123 4567"} {
		if strings.Contains(got, leak) {
			t.Errorf("value %q leaked into output:\n%s", leak, got)
		}
	}
	if !strings.Contains(got, "redacted_") {
		t.Errorf("expected redaction markers in output:\n%s", got)
	}
}

func TestHTML_DefaultHashesProse(t *testing.T) {
	out, err := redact.MakeDefaultPolicy().HTML([]byte(loginHTML))
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	got := string(out)
	if strings.Contains(got, "password") {
		t.Errorf("default should redact visible prose, but it leaked:\n%s", got)
	}
	if strings.Contains(got, "AbCdEf1234567890XyZ") {
		t.Errorf("default should redact the csrf token:\n%s", got)
	}
	if !strings.Contains(got, "<form") || !strings.Contains(got, "<input") {
		t.Errorf("tag structure should survive even under default:\n%s", got)
	}
}

func TestJSON_NumbersKeepExactText(t *testing.T) {
	out, err := relaxed().JSON([]byte(`{"id":998877665544332211,"ts":1726300000}`))
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if want := `{"id":"` + redact.MakeDefaultPolicy().String("998877665544332211") + `","ts":1726300000}`; string(out) != want {
		t.Errorf("hashed ID should match its string form and timestamp should survive exactly: got %s, want %s", out, want)
	}
}

func TestValue_UnsupportedTypesHashed(t *testing.T) {
	p := redact.MakeDefaultPolicy()
	for _, v := range []any{[]string{"secret"}, json.RawMessage(`"secret"`), int(1234567890), map[string]string{"k": "secret"}} {
		p.Value(&v)
		if got, _ := v.(string); !isRedacted(got) {
			t.Errorf("unsupported type should be hashed whole, got %T %v", v, v)
		}
	}
	for _, v := range []any{nil, true} {
		orig := v
		p.Value(&v)
		if v != orig {
			t.Errorf("%v should be kept, got %v", orig, v)
		}
	}
}

func TestString_MarkerGuardIsExact(t *testing.T) {
	p := redact.MakeDefaultPolicy()
	got := p.String(`{"a":"redacted_5char_str_deadbeef","secret":"supersecretvalue123"}`)
	if strings.Contains(got, "supersecretvalue123") {
		t.Errorf("a string merely containing a marker must still be redacted, got %q", got)
	}
	if got := p.String("my_redacted_token_AbCdEf123456"); !isRedacted(got) {
		t.Errorf("marker-like substring must not bypass redaction, got %q", got)
	}
}

func TestString_MarkerDeterministic(t *testing.T) {
	p := redact.MakeDefaultPolicy()
	const digitUUID = "12345678-1234-1234-1234-123456789012"
	want := p.String(digitUUID)
	if !strings.Contains(want, "uuid_lowercase") {
		t.Errorf("digit-only UUID should be typed uuid_lowercase, got %q", want)
	}
	for range 100 {
		if got := p.String(digitUUID); got != want {
			t.Fatalf("marker changed between calls: %q vs %q", got, want)
		}
	}
}

func TestHTML_AttributeKeys(t *testing.T) {
	const form = `<form method="post"><input type="hidden" class="ok"></form>`
	keep := redact.MakeDefaultPolicy()
	keep.ClassifyKey = func(key string) redact.Decision {
		if key == "type" || key == "method" {
			return redact.Keep
		}
		return redact.Auto
	}
	out, err := keep.HTML([]byte(form))
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	got := string(out)
	for _, want := range []string{`method="post"`, `type="hidden"`} {
		if !strings.Contains(got, want) {
			t.Errorf("ClassifyKey Keep should keep %s, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, `class="ok"`) {
		t.Errorf("unlisted attribute should still be hashed, got:\n%s", got)
	}

	hash := relaxed()
	hash.SensitiveKeys["class"] = true
	out, err = hash.HTML([]byte(form))
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}
	got = string(out)
	if strings.Contains(got, `class="ok"`) {
		t.Errorf("SensitiveKeys should apply to attribute names, got:\n%s", got)
	}
	if !strings.Contains(got, `type="hidden"`) {
		t.Errorf("relaxed policy should keep short attribute values, got:\n%s", got)
	}
}

func TestMakeDefaultPolicy_Independent(t *testing.T) {
	p := redact.MakeDefaultPolicy()
	p.SensitiveKeys["status"] = true
	if redact.MakeDefaultPolicy().SensitiveKeys["status"] {
		t.Error("mutating one policy's SensitiveKeys leaked into a fresh one")
	}
}

var redactedRegexp = regexp.MustCompile(`^redacted_\d+char_`)

func isRedacted(s string) bool {
	return redactedRegexp.MatchString(s)
}
