package progress

import (
	"fmt"
	"io"
)

// Reader is an [io.ReadSeekCloser] that reports the number of bytes read from it via a callback.
//
// The callback is called at most every "updateInterval" bytes.
// The updateInterval can be set using the [Reader.WithUpdateInterval] method.
// The callback will also be called whenever [Reader.Seek] is called.
//
// The following is an example of how to use [Reader] to report the progress of
// reading from a file:
//
//	file, _ := os.Open("file.txt")
//	progressReader := NewReader(f, func(readBytes int) {
//		fmt.Printf("Read %d bytes\n", readBytes)
//	})
//	io.ReadAll(progressReader)
type Reader struct {
	inner          io.Reader
	readBytes      int
	progressFn     func(readBytes int)
	lastUpdate     int
	updateInterval int
}

func NewReader(r io.Reader, progressFn func(readBytes int)) *Reader {
	return &Reader{inner: r, progressFn: progressFn, updateInterval: defaultUpdateInterval}
}

func (r *Reader) WithUpdateInterval(bytes int) *Reader {
	r.updateInterval = bytes
	return r
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.inner.Read(p)
	if err != nil {
		return n, err
	}
	r.readBytes += n
	if r.lastUpdate == 0 || r.readBytes-r.lastUpdate > r.updateInterval {
		r.progressFn(r.readBytes)
		r.lastUpdate = r.readBytes
	}
	return n, nil
}

func (r *Reader) Close() error {
	if closer, ok := r.inner.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := r.inner.(io.ReadSeeker)
	if !ok {
		return 0, fmt.Errorf("progress.Reader: source reader (%T) is not an io.ReadSeeker", r.inner)
	}
	n, err := seeker.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	r.readBytes = int(n)
	r.progressFn(r.readBytes)
	r.lastUpdate = r.readBytes
	return n, nil
}

var _ io.ReadSeekCloser = (*Reader)(nil)
