// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package l10n

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	badGlobalLog "github.com/rs/zerolog/log" // zerolog-allow-global-log
)

var (
	namespacePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*$`)
	idPattern          = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)
	placeholderPattern = regexp.MustCompile(`\{([a-z][a-z0-9_]*)\}`)
)

var (
	registryLock sync.Mutex
	templates    = map[string]*Template{}
)

// Namespace is the key prefix for one project's templates, e.g. "mautrix-whatsapp".
type Namespace string

// NewNamespace validates the name and returns it as a Namespace.
func NewNamespace(name string) Namespace {
	if !namespacePattern.MatchString(name) {
		panic(fmt.Errorf("l10n: invalid namespace %q", name))
	}
	return Namespace(name)
}

// Template is an English message with {name} placeholders. Declare templates as
// package-level variables so they're registered before the catalog is generated.
type Template struct {
	key         string
	english     string
	englishOne  string
	pluralParam string
	params      []string
	comment     string
}

type Option func(*Template)

// WithComment adds a note for translators.
func WithComment(comment string) Option {
	return func(t *Template) {
		t.comment = comment
	}
}

// Define registers a template. It panics if the key is already registered.
func (ns Namespace) Define(id, english string, opts ...Option) *Template {
	t := &Template{english: english}
	t.params = placeholders(english)
	return ns.register(id, t, opts)
}

// DefinePlural registers a template that uses one when countParam is 1 and other otherwise.
func (ns Namespace) DefinePlural(id, countParam, one, other string, opts ...Option) *Template {
	t := &Template{english: other, englishOne: one, pluralParam: countParam}
	t.params = placeholders(one, other)
	if !slices.Contains(t.params, countParam) {
		t.params = append(t.params, countParam)
		slices.Sort(t.params)
	}
	return ns.register(id, t, opts)
}

func (ns Namespace) register(id string, t *Template, opts []Option) *Template {
	if !idPattern.MatchString(id) {
		panic(fmt.Errorf("l10n: invalid template ID %q in namespace %s", id, ns))
	}
	t.key = string(ns) + ":" + id
	for _, opt := range opts {
		opt(t)
	}
	registryLock.Lock()
	defer registryLock.Unlock()
	if _, exists := templates[t.key]; exists {
		panic(fmt.Errorf("l10n: duplicate template key %s", t.key))
	}
	templates[t.key] = t
	return t
}

func placeholders(texts ...string) []string {
	var names []string
	for _, text := range texts {
		for _, match := range placeholderPattern.FindAllStringSubmatch(text, -1) {
			if !slices.Contains(names, match[1]) {
				names = append(names, match[1])
			}
		}
	}
	slices.Sort(names)
	return names
}

func (t *Template) Key() string {
	return t.key
}

// RenderErrorHandler is called when Render gets params that don't match the
// template. The rendered message is then unlocalized. The default panics in
// tests and logs otherwise.
var RenderErrorHandler = func(t *Template, err error) {
	if testing.Testing() {
		panic(err)
	}
	badGlobalLog.Warn().Err(err).Str("l10n_key", t.key).Msg("Failed to render localizable message")
}

// Message renders a template that has no params.
func (t *Template) Message() Message {
	return t.Render(nil)
}

// Render fills in the template's placeholders. Values must be strings, integers or Messages.
func (t *Template) Render(params Params) Message {
	normalized, err := t.normalizeParams(params)
	english := t.english
	if t.pluralParam != "" {
		if count, ok := normalized[t.pluralParam].(int64); ok && count == 1 {
			english = t.englishOne
		}
	}
	text := placeholderPattern.ReplaceAllStringFunc(english, func(placeholder string) string {
		name := placeholder[1 : len(placeholder)-1]
		val, ok := normalized[name]
		if !ok {
			return placeholder
		}
		return fmt.Sprint(val)
	})
	if err != nil {
		RenderErrorHandler(t, err)
		return Unlocalized(text)
	}
	return Message{Key: t.key, Text: text, Params: normalized}
}

func (t *Template) normalizeParams(params Params) (Params, error) {
	var errs []string
	var normalized Params
	if len(params) > 0 {
		normalized = make(Params, len(params))
	}
	for name, val := range params {
		if !slices.Contains(t.params, name) {
			errs = append(errs, fmt.Sprintf("unknown param %q", name))
			continue
		}
		if msg, ok := val.(*Message); ok && msg != nil {
			val = *msg
		}
		if msg, ok := val.(Message); ok {
			normalized[name] = msg
			continue
		}
		rv := reflect.ValueOf(val)
		_, isStringer := val.(fmt.Stringer)
		_, isError := val.(error)
		switch kind := rv.Kind(); {
		case (isStringer || isError) && kind != reflect.String:
			// e.g. time.Duration would otherwise render as a bare number
			errs = append(errs, fmt.Sprintf("param %q has unsupported type %T", name, val))
			normalized[name] = fmt.Sprint(val)
		case kind == reflect.String:
			normalized[name] = rv.String()
		case kind >= reflect.Int && kind <= reflect.Int64:
			normalized[name] = rv.Int()
		case kind >= reflect.Uint8 && kind <= reflect.Uint32:
			normalized[name] = int64(rv.Uint())
		default:
			errs = append(errs, fmt.Sprintf("param %q has unsupported type %T", name, val))
			normalized[name] = fmt.Sprint(val)
		}
	}
	for _, name := range t.params {
		if _, ok := params[name]; !ok {
			errs = append(errs, fmt.Sprintf("missing param %q", name))
		}
	}
	if t.pluralParam != "" {
		if _, ok := normalized[t.pluralParam].(int64); !ok {
			errs = append(errs, fmt.Sprintf("plural param %q must be an integer", t.pluralParam))
		}
	}
	if len(errs) > 0 {
		slices.Sort(errs)
		return normalized, fmt.Errorf("l10n: invalid params for %s: %s", t.key, strings.Join(errs, ", "))
	}
	return normalized, nil
}

// CatalogEntry is one template in a translation catalog.
type CatalogEntry struct {
	Key string `json:"key"`
	// English is the "other" form for plural templates.
	English     string   `json:"english"`
	EnglishOne  string   `json:"english_one,omitempty"`
	PluralParam string   `json:"plural_param,omitempty"`
	Params      []string `json:"params,omitempty"`
	Comment     string   `json:"comment,omitempty"`
}

// Catalog returns the namespace's registered templates sorted by key.
func Catalog(ns Namespace) []CatalogEntry {
	return catalog(string(ns) + ":")
}

// CatalogAll returns every registered template sorted by key.
func CatalogAll() []CatalogEntry {
	return catalog("")
}

func catalog(prefix string) []CatalogEntry {
	registryLock.Lock()
	defer registryLock.Unlock()
	var entries []CatalogEntry
	for key, t := range templates {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		entries = append(entries, CatalogEntry{
			Key:         t.key,
			English:     t.english,
			EnglishOne:  t.englishOne,
			PluralParam: t.pluralParam,
			Params:      t.params,
			Comment:     t.comment,
		})
	}
	slices.SortFunc(entries, func(a, b CatalogEntry) int {
		return strings.Compare(a.Key, b.Key)
	})
	return entries
}
