//go:build go1.21

package exslices_test

import (
	"cmp"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/util/exslices"
)

func generateInts(seed int64, size, maxVal int, sorted bool) ([]int, []int) {
	l1 := make([]int, size)
	l2 := make([]int, size)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < size; i++ {
		l1[i] = r.Intn(maxVal)
		l2[i] = r.Intn(maxVal)
	}
	if sorted {
		sort.Ints(l1)
		sort.Ints(l2)
		l1 = slices.Compact(l1)
		l2 = slices.Compact(l2)
	}
	return l1, l2
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(r *rand.Rand, length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(alphabet[r.Intn(len(alphabet))])
	}
	return sb.String()
}

func generateStrings(seed int64, size, length int, sorted bool) ([]string, []string) {
	l1 := make([]string, size)
	l2 := make([]string, size)
	r := rand.New(rand.NewSource(seed))
	for i := 0; i < size; i++ {
		l1[i] = randomString(r, length)
		l2[i] = randomString(r, length)
	}
	if sorted {
		sort.Strings(l1)
		sort.Strings(l2)
		l1 = slices.Compact(l1)
		l2 = slices.Compact(l2)
	}
	return l1, l2
}

func TestDiff_Ints_Basic(t *testing.T) {
	l1 := []int{1, 2, 3, 4, 5}
	l2 := []int{3, 4, 5, 6, 7}
	uniqueToA, uniqueToB := exslices.Diff(l1, l2)
	sort.Ints(uniqueToA)
	sort.Ints(uniqueToB)
	assert.Equal(t, []int{1, 2}, uniqueToA)
	assert.Equal(t, []int{6, 7}, uniqueToB)
}

func TestSortedDiff_Ints_Edge(t *testing.T) {
	testCases := []struct {
		name         string
		l1, l2, a, b []int
	}{
		{
			"shorter left",
			[]int{1, 2}, []int{1, 3, 4, 5},
			[]int{2}, []int{3, 4, 5},
		},
		{
			"shorter right",
			[]int{1, 3, 4, 5}, []int{1, 2},
			[]int{3, 4, 5}, []int{2},
		},
		{
			"empty side",
			[]int{}, []int{1, 2, 3},
			[]int{}, []int{1, 2, 3},
		},
		{
			"empty both",
			[]int{}, []int{},
			[]int{}, []int{},
		},
		{
			"equal",
			[]int{1, 2, 3}, []int{1, 2, 3},
			[]int{}, []int{},
		},
		{
			"jump",
			[]int{1, 999, 1001}, []int{1, 2, 3, 4, 1000, 1001},
			[]int{999}, []int{2, 3, 4, 1000},
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			uniqueToA, uniqueToB := exslices.Diff(c.l1, c.l2)
			sort.Ints(uniqueToA)
			sort.Ints(uniqueToB)
			assert.Equal(t, c.a, uniqueToA)
			assert.Equal(t, c.b, uniqueToB)
		})
	}
}

func TestDiff_Strings_Basic(t *testing.T) {
	l1 := []string{"a", "b", "c", "d", "e"}
	l2 := []string{"c", "d", "e", "f", "g"}
	uniqueToA, uniqueToB := exslices.Diff(l1, l2)
	sort.Strings(uniqueToA)
	sort.Strings(uniqueToB)
	assert.Equal(t, []string{"a", "b"}, uniqueToA)
	assert.Equal(t, []string{"f", "g"}, uniqueToB)
}

func TestSortedDiff_Strings_Basic(t *testing.T) {
	l1 := []string{"a", "b", "c", "d", "e"}
	l2 := []string{"c", "d", "e", "f", "g"}
	uniqueToA, uniqueToB := exslices.SortedDiff(l1, l2, cmp.Compare[string])
	assert.Equal(t, []string{"a", "b"}, uniqueToA)
	assert.Equal(t, []string{"f", "g"}, uniqueToB)
}

func tSortedDiff_Random[T cmp.Ordered](t *testing.T, generateFunc func(int64, int, int, bool) ([]T, []T)) {
	l1, l2 := generateFunc(1, 1000, 1000, true)
	uniqueToA1, uniqueToB1 := exslices.SortedDiff(l1, l2, cmp.Compare[T])
	uniqueToA2, uniqueToB2 := exslices.Diff(l1, l2)
	// Diff doesn't guarantee order, so sort the results
	slices.Sort(uniqueToA2)
	slices.Sort(uniqueToB2)
	assert.Equal(t, uniqueToA2, uniqueToA1)
	assert.Equal(t, uniqueToB2, uniqueToB1)
}

func TestSortedDiff_Random(t *testing.T) {
	t.Run("Ints", func(t *testing.T) {
		tSortedDiff_Random[int](t, generateInts)
	})
	t.Run("Strings", func(t *testing.T) {
		tSortedDiff_Random[string](t, generateStrings)
	})
}

func bDiff[T cmp.Ordered](b *testing.B, generateFunc func(int64, int, int, bool) ([]T, []T), sorted bool) {
	l1, l2 := generateFunc(1, 1000, 1000, sorted)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exslices.Diff(l1, l2)
	}
}

func BenchmarkDiff(b *testing.B) {
	b.Run("Ints", func(b *testing.B) {
		bDiff(b, generateInts, false)
	})
	b.Run("Ints/Sorted", func(b *testing.B) {
		bDiff(b, generateInts, true)
	})
	b.Run("Strings", func(b *testing.B) {
		bDiff(b, generateStrings, false)
	})
	b.Run("Strings/Sorted", func(b *testing.B) {
		bDiff(b, generateStrings, true)
	})
}

func bSortedDiff[T comparable](b *testing.B, generateFunc func(int64, int, int, bool) ([]T, []T), compareFunc func(a, b T) int) {
	l1, l2 := generateFunc(1, 1000, 1000, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exslices.SortedDiff(l1, l2, compareFunc)
	}
}

func BenchmarkSortedDiff(b *testing.B) {
	b.Run("Ints", func(b *testing.B) {
		bSortedDiff(b, generateInts, cmp.Compare[int])
	})
	b.Run("Strings", func(b *testing.B) {
		bSortedDiff(b, generateStrings, cmp.Compare[string])
	})
}
