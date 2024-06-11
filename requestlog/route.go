package requestlog

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

type Route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc

	TrackHTTPMetrics func(*Route) func(*CountingResponseWriter)

	LogContent bool
}

var _ http.Handler = (*Route)(nil)

func (rt *Route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	crw := w.(*CountingResponseWriter)
	if rt.TrackHTTPMetrics != nil {
		defer rt.TrackHTTPMetrics(rt)(crw)
	}
	if rt.LogContent {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			crw.ResponseBody = &bytes.Buffer{}
		}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			pcr := &partialCachingReader{Reader: r.Body}
			crw.RequestBody = &pcr.Buffer
			r.Body = pcr
		}
	}
	rt.Handler(w, r)
}

type partialCachingReader struct {
	Reader io.ReadCloser
	Buffer bytes.Buffer
}

func (pcr *partialCachingReader) Read(p []byte) (int, error) {
	n, err := pcr.Reader.Read(p)
	if n > 0 {
		pcr.Buffer.Write(CutRequestData(p[:n], pcr.Buffer.Len()))
	}
	return n, err
}

func (pcr *partialCachingReader) Close() error {
	return pcr.Reader.Close()
}
