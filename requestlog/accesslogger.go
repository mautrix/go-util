package requestlog

import (
	"bytes"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const MaxRequestSizeLog = 4 * 1024
const MaxStringRequestSizeLog = MaxRequestSizeLog / 2

type Options struct {
	// Should OPTIONS requests be logged?
	LogOptions bool
	// Should remote_addr logging prefer X-Forwarded-For if present?
	TrustXForwardedFor bool
	// Should we recover from panics?
	Recover bool
}

func AccessLogger(opts Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := hlog.FromRequest(r)

			crw := &CountingResponseWriter{
				ResponseWriter: w,
				ResponseLength: -1,
				StatusCode:     -1,
			}

			start := time.Now()

			fillRequestLog := func(requestLog *zerolog.Event) {
				requestDuration := time.Since(start)

				if userAgent := r.UserAgent(); userAgent != "" {
					requestLog.Str("user_agent", userAgent)
				}
				if referer := r.Referer(); referer != "" {
					requestLog.Str("referer", referer)
				}
				remoteAddr := r.RemoteAddr
				if opts.TrustXForwardedFor {
					forwarded := strings.Split(r.Header.Get("X-Forwarded-For"), ", ")
					if len(forwarded) > 0 && len(forwarded[0]) > 0 {
						requestLog.Str("x_forwarded_for", forwarded[0])
					}
				}

				requestLog.Str("remote_addr", remoteAddr)
				requestLog.Str("method", r.Method)
				requestLog.Str("proto", r.Proto)
				requestLog.Int64("request_length", r.ContentLength)
				requestLog.Str("host", r.Host)
				requestLog.Str("request_uri", r.RequestURI)
				if r.Method != http.MethodGet && r.Method != http.MethodHead {
					requestLog.Str("request_content_type", r.Header.Get("Content-Type"))
					if crw.RequestBody != nil {
						logRequestMaybeJSON(requestLog, "request_body", crw.RequestBody.Bytes())
					}
				}

				// response
				requestLog.Int64("request_time_ms", requestDuration.Milliseconds())
				requestLog.Int("status_code", crw.StatusCode)
				requestLog.Int("response_length", crw.ResponseLength)
				requestLog.Str("response_content_type", crw.Header().Get("Content-Type"))
				if crw.ResponseBody != nil {
					logRequestMaybeJSON(requestLog, "response_body", crw.ResponseBody.Bytes())
				}
			}

			if opts.Recover {
				defer func() {
					if rvr := recover(); rvr != nil {
						if rvr == http.ErrAbortHandler {
							panic(rvr)
						}

						if crw.StatusCode == -1 && r.Header.Get("Connection") != "Upgrade" {
							crw.StatusCode = http.StatusInternalServerError
							w.WriteHeader(crw.StatusCode)
						}

						requestLog := log.Error()
						fillRequestLog(requestLog)

						requestLog.Bytes(zerolog.ErrorStackFieldName, debug.Stack())
						if err, ok := rvr.(error); ok {
							requestLog.Err(err)
						} else {
							requestLog.Any(zerolog.ErrorFieldName, rvr)
						}
						requestLog.Msg("Access")
					}
				}()
			}

			next.ServeHTTP(crw, r)

			if r.Method == http.MethodOptions && !opts.LogOptions {
				return
			}

			// don't log successful health requests
			if r.URL.Path == "/health" && crw.StatusCode == http.StatusNoContent {
				return
			}

			var requestLog *zerolog.Event
			if crw.StatusCode >= 500 {
				requestLog = log.Error()
			} else if crw.StatusCode >= 400 {
				requestLog = log.Warn()
			} else {
				requestLog = log.Info()
			}

			fillRequestLog(requestLog)
			requestLog.Msg("Access")
		})
	}
}

func logRequestMaybeJSON(evt *zerolog.Event, key string, data []byte) {
	data = removeNewlines(data)
	if json.Valid(data) {
		evt.RawJSON(key, data)
	} else {
		// Logging as a string will create lots of escaping and it's not valid json anyway, so cut off a bit more
		if len(data) > MaxStringRequestSizeLog {
			data = data[:MaxStringRequestSizeLog]
		}
		evt.Bytes(key+"_invalid", data)
	}
}

func removeNewlines(data []byte) []byte {
	data = bytes.TrimSpace(data)
	if bytes.ContainsRune(data, '\n') {
		data = bytes.ReplaceAll(data, []byte{'\n'}, []byte{})
		data = bytes.ReplaceAll(data, []byte{'\r'}, []byte{})
	}
	return data
}
