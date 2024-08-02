// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package confusable_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/util/confusable"
)

func TestSkeleton(t *testing.T) {
	assert.Equal(t, "MEOW MEOW", confusable.Skeleton("MEOW 𝗠𝔼𝑶𝓦"))
}

func TestConfusable(t *testing.T) {
	assert.True(t, confusable.Confusable("MEOW", "𝗠𝔼𝑶𝓦"))
}

func BenchmarkSkeleton(b *testing.B) {
	for i := 0; i < b.N; i++ {
		confusable.Skeleton("MEOW ⋘ 𝗠𝔼𝑶𝓦 MEOW ⋘ 𝗠𝔼𝑶𝓦")
	}
}

func BenchmarkSkeletonBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		confusable.SkeletonBytes("MEOW ⋘ 𝗠𝔼𝑶𝓦 MEOW ⋘ 𝗠𝔼𝑶𝓦")
	}
}
