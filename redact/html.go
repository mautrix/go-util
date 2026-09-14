// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package redact

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// HTML redacts an HTML document, keeping the tag and attribute structure while
// running text content, comments and attribute values through the policy.
// Attribute values are subject to the same key rules as object values, keyed
// by attribute name. A <script type="application/json"> body is valid JSON, so
// it is redacted structurally; other script and style bodies are treated as
// opaque strings.
func (p Policy) HTML(data []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML for redaction: %w", err)
	}
	p.redactNode(doc)
	var buf bytes.Buffer
	if err = html.Render(&buf, doc); err != nil {
		return nil, fmt.Errorf("rendering redacted HTML: %w", err)
	}
	return buf.Bytes(), nil
}

func (p Policy) redactNode(n *html.Node) {
	switch n.Type {
	case html.TextNode, html.CommentNode:
		if strings.TrimSpace(n.Data) != "" {
			n.Data = p.String(n.Data)
		}
	case html.ElementNode:
		for i, attr := range n.Attr {
			n.Attr[i].Val = p.keyed(attr.Key, attr.Val).(string)
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		p.redactNode(child)
	}
}
