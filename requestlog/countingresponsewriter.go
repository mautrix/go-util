package requestlog

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type CountingResponseWriter struct {
	StatusCode     int
	ResponseLength int
	Hijacked       bool
	ResponseWriter http.ResponseWriter
	ResponseBody   *bytes.Buffer
	RequestBody    *bytes.Buffer
}

var (
	_ http.ResponseWriter = (*CountingResponseWriter)(nil)
	_ http.Flusher        = (*CountingResponseWriter)(nil)
	_ http.Hijacker       = (*CountingResponseWriter)(nil)
)

func (crw *CountingResponseWriter) Header() http.Header {
	return crw.ResponseWriter.Header()
}

func (crw *CountingResponseWriter) Write(data []byte) (int, error) {
	if crw.ResponseLength == -1 {
		crw.ResponseLength = 0
	}
	if crw.StatusCode == -1 {
		crw.StatusCode = http.StatusOK
	}
	crw.ResponseLength += len(data)

	if crw.ResponseBody != nil && crw.ResponseBody.Len() < MaxRequestSizeLog {
		crw.ResponseBody.Write(CutRequestData(data, crw.ResponseBody.Len()))
	}
	return crw.ResponseWriter.Write(data)
}

func (crw *CountingResponseWriter) WriteHeader(statusCode int) {
	crw.StatusCode = statusCode
	crw.ResponseWriter.WriteHeader(statusCode)
	if !strings.HasPrefix(crw.Header().Get("Content-Type"), "application/json") {
		crw.ResponseBody = nil
	}
}

func (crw *CountingResponseWriter) Flush() {
	flusher, ok := crw.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

func (crw *CountingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := crw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("CountingResponseWriter: %T does not implement http.Hijacker", crw.ResponseWriter)
	}
	crw.Hijacked = true
	return hijacker.Hijack()
}

func CutRequestData(data []byte, length int) []byte {
	if len(data)+length > MaxRequestSizeLog {
		return data[:MaxRequestSizeLog-length]
	}
	return data
}
