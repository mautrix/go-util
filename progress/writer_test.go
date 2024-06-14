package progress_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mau.fi/util/progress"
)

func TestWriter(t *testing.T) {
	var progressUpdates []int
	progressWriter := progress.NewWriter(func(processedBytes int) {
		progressUpdates = append(progressUpdates, processedBytes)
	})

	for i := 0; i < 10; i++ {
		_, err := io.Copy(progressWriter, bytes.NewReader(bytes.Repeat([]byte{42}, 256*1024)))
		require.NoError(t, err)
	}

	assert.Greater(t, len(progressUpdates), 1)
	assert.IsIncreasing(t, progressUpdates)
}
