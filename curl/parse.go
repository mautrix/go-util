// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package curl

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"go.mau.fi/util/shlex"
)

type Parsed struct {
	*http.Request
	ParsedJSON map[string]any
}

var ansiCQuoteReplacer = strings.NewReplacer(
	`\n`, "\n",
	`\r`, "\r",
	`\t`, "\t",
	`\v`, "\v",
	`\a`, "\a",
	`\b`, "\b",
	`\f`, "\f",
	`\f`, "\f",
	`\f`, "\f",
	`\?`, "?",
	`\'`, "'",
	`\"`, `"`,
	`\\`, `\`,
)

func Parse(curl string) (*Parsed, error) {
	parts, err := shlex.Split(curl)
	if err != nil {
		return nil, fmt.Errorf("failed to split command: %w", err)
	} else if parts[0] != "curl" {
		return nil, fmt.Errorf("expected command to start with curl, got %q", parts[0])
	}
	req := &http.Request{
		Header: make(http.Header),
	}
	var body string
	for i := 1; i < len(parts); i++ {
		val := parts[i]
		if req.URL == nil && strings.HasPrefix(val, "https://") {
			req.URL, err = url.Parse(val)
			if err != nil {
				return nil, fmt.Errorf("failed to parse url: %w", err)
			}
		}
		switch val {
		case "-H":
			i++
			hdrParts := strings.SplitN(parts[i], ": ", 2)
			req.Header.Add(hdrParts[0], hdrParts[1])
		case "--data-raw", "--data-binary":
			i++
			body = parts[i]
			if strings.HasPrefix(body, "$") {
				body = ansiCQuoteReplacer.Replace(body[1:])
			}
		case "-X":
			i++
			req.Method = parts[i]
		case "-b":
			i++
			req.Header.Add("Cookie", parts[i])
		}
	}
	req.Body = io.NopCloser(strings.NewReader(body))
	contentType := req.Header.Get("Content-Type")
	var jsonBody map[string]any
	if contentType != "" {
		var params map[string]string
		contentType, params, err = mime.ParseMediaType(contentType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content type: %w", err)
		}
		switch contentType {
		case "application/json":
			err = json.Unmarshal([]byte(body), &jsonBody)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON body: %w", err)
			}
		case "multipart/form-data":
			req.MultipartForm, err = multipart.NewReader(strings.NewReader(body), params["boundary"]).ReadForm(1024 * 1024)
			if err != nil {
				return nil, fmt.Errorf("failed to parse form data: %w", err)
			}
		}
	}
	return &Parsed{
		Request:    req,
		ParsedJSON: jsonBody,
	}, nil
}
