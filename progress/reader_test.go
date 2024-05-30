package progress_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mau.fi/util/progress"
)

func TestReader(t *testing.T) {
	reader := bytes.NewReader(bytes.Repeat([]byte{42}, 1024*1024))

	var progressUpdates []int
	progressReader := progress.NewReader(reader, func(readBytes int) {
		progressUpdates = append(progressUpdates, readBytes)
	})

	data, err := io.ReadAll(progressReader)
	assert.NoError(t, err)
	assert.Equal(t, data, bytes.Repeat([]byte{42}, 1024*1024))

	assert.Greater(t, len(progressUpdates), 1)
	assert.IsIncreasing(t, progressUpdates)
}

type testReader struct {
	*bytes.Reader
	closed bool
}

func (r *testReader) Close() error {
	if r.closed {
		return errors.New("already closed")
	}
	r.closed = true
	return nil
}

func TestReadCloser(t *testing.T) {
	readCloser := &testReader{Reader: bytes.NewReader(bytes.Repeat([]byte{42}, 1024*1024))}

	var progressUpdates []int
	progressReader := progress.NewReader(readCloser, func(readBytes int) {
		progressUpdates = append(progressUpdates, readBytes)
	})

	data, err := io.ReadAll(progressReader)
	assert.NoError(t, err)
	assert.Equal(t, data, bytes.Repeat([]byte{42}, 1024*1024))

	assert.Greater(t, len(progressUpdates), 1)
	assert.IsIncreasing(t, progressUpdates)

	assert.NoError(t, progressReader.Close())
	assert.ErrorContains(t, progressReader.Close(), "already closed")
}
