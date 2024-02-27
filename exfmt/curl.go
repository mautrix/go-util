// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exfmt

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// FormatCurl formats the given HTTP request as a curl command.
//
// This will include all headers, and also the request body if GetBody is set. Notes:
//
// * Header names are quoted using fmt.Sprintf, so it may not always be correct for shell quoting.
// * The URL is only quoted and not escaped, so URLs with single quotes will not currently work.
//
// The client parameter is optional and is used to find cookies from the cookie jar.
func FormatCurl(cli *http.Client, req *http.Request) string {
	var curl []string
	hasBody := false
	if req.GetBody != nil {
		body, _ := req.GetBody()
		if body != http.NoBody {
			b, _ := io.ReadAll(body)
			curl = []string{"echo", base64.StdEncoding.EncodeToString(b), "|", "base64", "-d", "|"}
			hasBody = true
		}
	}
	curl = append(curl, "curl", "-v")
	switch req.Method {
	case http.MethodGet:
		// Don't add -X
	case http.MethodHead:
		curl = append(curl, "-I")
	default:
		curl = append(curl, "-X", req.Method)
	}
	for key, vals := range req.Header {
		kv := fmt.Sprintf("%s: %s", key, vals[0])
		curl = append(curl, "-H", fmt.Sprintf("%q", kv))
	}
	if cli != nil && cli.Jar != nil {
		cookies := cli.Jar.Cookies(req.URL)
		if len(cookies) > 0 {
			cookieStrings := make([]string, len(cookies))
			for i, cookie := range cookies {
				if strings.ContainsAny(cookie.Value, " ,") {
					cookieStrings[i] = fmt.Sprintf(`%s="%s"`, cookie.Name, cookie.Value)
				} else {
					cookieStrings[i] = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
				}
			}
			curl = append(curl, "-H", fmt.Sprintf("%q", "Cookie: "+strings.Join(cookieStrings, "; ")))
		}
	}
	if hasBody {
		curl = append(curl, "--data-binary", "@-")
	}
	curl = append(curl, fmt.Sprintf("'%s'", req.URL.String()))
	return strings.Join(curl, " ")
}
