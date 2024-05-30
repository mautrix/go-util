package progress

import "io"

// Writer is an [io.Writer] that reports the number of bytes written to it via
// a callback. The callback is called at most every "updateInterval" bytes. The
// updateInterval can be set using the [Writer.WithUpdateInterval] method.
//
// The following is an example of how to use [Writer] to report the progress of
// writing to a file:
//
//	file, _ := os.Create("file.txt")
//	progressWriter := progress.NewWriter(func(processedBytes int) {
//	    fmt.Printf("Processed %d bytes\n", processedBytes)
//	})
//	writerWithProgress := io.MultiWriter(file, progressWriter)
//	io.Copy(writerWithProgress, bytes.NewReader(bytes.Repeat([]byte{42}, 1024*1024)))
type Writer struct {
	processedBytes int
	progressFn     func(processedBytes int)
	lastUpdate     int
	updateInterval int
}

func NewWriter(progressFn func(processedBytes int)) *Writer {
	return &Writer{progressFn: progressFn, updateInterval: defaultUpdateInterval}
}

func (w *Writer) WithUpdateInterval(bytes int) *Writer {
	w.updateInterval = bytes
	return w
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.processedBytes += len(p)
	if w.lastUpdate == 0 || w.processedBytes-w.lastUpdate > w.updateInterval {
		w.progressFn(w.processedBytes)
		w.lastUpdate = w.processedBytes
	}
	return len(p), nil
}

var _ io.Writer = (*Writer)(nil)
