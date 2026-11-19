// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package l10ntest checks committed translation catalogs against registered templates.
package l10ntest

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.mau.fi/util/l10n"
)

// UpdateEnv makes CheckCatalog rewrite the file instead of failing when set to 1.
const UpdateEnv = "L10N_UPDATE"

type CatalogFile struct {
	Namespace string              `json:"namespace"`
	Messages  []l10n.CatalogEntry `json:"messages"`
}

// GenerateCatalog returns the catalog file contents for a namespace.
func GenerateCatalog(ns l10n.Namespace) ([]byte, error) {
	data, err := json.MarshalIndent(&CatalogFile{
		Namespace: string(ns),
		Messages:  l10n.Catalog(ns),
	}, "", "\t")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// CheckCatalog fails if the file at path doesn't match the namespace's templates.
// The calling package must import every package that defines templates.
func CheckCatalog(t *testing.T, path string, ns l10n.Namespace) {
	t.Helper()
	expected, err := GenerateCatalog(ns)
	if err != nil {
		t.Fatalf("failed to generate catalog: %v", err)
	}
	if os.Getenv(UpdateEnv) == "1" {
		if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create catalog directory: %v", err)
		} else if err = os.WriteFile(path, expected, 0o644); err != nil {
			t.Fatalf("failed to write catalog: %v", err)
		}
		return
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read catalog (run with %s=1 to create it): %v", UpdateEnv, err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("%s is out of date, run the test with %s=1 to update it", path, UpdateEnv)
	}
}
