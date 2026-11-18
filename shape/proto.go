// Copyright (C) 2026 Nick Mills-Barrett
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package shape

import (
	"bytes"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	maxProtoDepth    = 8
	maxProtoFields   = 32
	maxProtoLen      = 2048
	maxProtoVarint   = 255
	maxProtoFieldNum = 10000
)

// ProtoFields returns the sorted numbers of the known fields that are set on a
// message.
func ProtoFields(msg proto.Message) []int32 {
	if msg == nil {
		return nil
	}
	var fields []int32
	msg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		fields = append(fields, int32(fd.Number()))
		return true
	})
	slices.Sort(fields)
	return fields
}

// ProtoUnknownFields returns the sorted, distinct numbers of the fields on a
// message that its descriptor doesn't know.
func ProtoUnknownFields(msg proto.Message) []int32 {
	if msg == nil {
		return nil
	}
	var fields []int32
	unknown := msg.ProtoReflect().GetUnknown()
	for len(unknown) > 0 {
		num, typ, n := protowire.ConsumeTag(unknown)
		if n < 0 {
			break
		}
		unknown = unknown[n:]
		n = protowire.ConsumeFieldValue(num, typ, unknown)
		if n < 0 {
			break
		}
		unknown = unknown[n:]
		fields = append(fields, int32(num))
	}
	slices.Sort(fields)
	return slices.Compact(fields)
}

// Protowire describes raw protobuf wire data without a schema, e.g.
// "{1:v=3,2:b16:id,4:b12{1:f64,2:v}}". Each field is its number and wire type:
// v for varints, with the value when it's at most 255, f32 and f64 for fixed
// values, and b plus the length for bytes. A bytes field is followed by
// :x and its hex for a single non-printable byte, :id for a 16-64 character
// token, :url for text containing a URL, :txt for other printable text, or its
// own shape in braces when it parses as a message. Wire data that doesn't parse
// as a message is described as a bytes value. At most 32 fields are described
// per message and 8 messages deep, and output over 2048 bytes is cut off.
func Protowire(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var sb strings.Builder
	if isToken(data) || isPrintable(data) || !isWellFormedProto(data) {
		sb.WriteByte('b')
		sb.WriteString(strconv.Itoa(len(data)))
		writeProtoBytes(&sb, data, maxProtoDepth)
		return sb.String()
	}
	writeProtoMessage(&sb, data, 0)
	if sb.Len() > maxProtoLen {
		return sb.String()[:maxProtoLen] + "..."
	}
	return sb.String()
}

func writeProtoMessage(sb *strings.Builder, data []byte, depth int) {
	sb.WriteByte('{')
	defer sb.WriteByte('}')
	for fields := 0; len(data) > 0; fields++ {
		if fields >= maxProtoFields || sb.Len() > maxProtoLen {
			sb.WriteString(",...")
			return
		} else if fields > 0 {
			sb.WriteByte(',')
		}
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			sb.WriteString("invalid")
			return
		}
		data = data[n:]
		sb.WriteString(strconv.Itoa(int(num)))
		sb.WriteByte(':')
		switch typ {
		case protowire.VarintType:
			var value uint64
			value, n = protowire.ConsumeVarint(data)
			if n >= 0 {
				sb.WriteByte('v')
				if value <= maxProtoVarint {
					sb.WriteByte('=')
					sb.WriteString(strconv.FormatUint(value, 10))
				}
			}
		case protowire.Fixed32Type:
			_, n = protowire.ConsumeFixed32(data)
			if n >= 0 {
				sb.WriteString("f32")
			}
		case protowire.Fixed64Type:
			_, n = protowire.ConsumeFixed64(data)
			if n >= 0 {
				sb.WriteString("f64")
			}
		case protowire.BytesType:
			var value []byte
			value, n = protowire.ConsumeBytes(data)
			if n >= 0 {
				sb.WriteByte('b')
				sb.WriteString(strconv.Itoa(len(value)))
				writeProtoBytes(sb, value, depth)
			}
		default:
			sb.WriteString("wt")
			sb.WriteString(strconv.Itoa(int(typ)))
			return
		}
		if n < 0 {
			sb.WriteString("truncated")
			return
		}
		data = data[n:]
	}
}

func writeProtoBytes(sb *strings.Builder, value []byte, depth int) {
	switch {
	case len(value) == 0:
	case len(value) == 1 && (value[0] < 0x20 || value[0] >= 0x7f):
		sb.WriteString(":x")
		sb.WriteString(hex.EncodeToString(value))
	case isToken(value):
		sb.WriteString(":id")
	case bytes.Contains(value, []byte("https://")) || bytes.Contains(value, []byte("http://")):
		sb.WriteString(":url")
	case isPrintable(value):
		sb.WriteString(":txt")
	case depth < maxProtoDepth && isWellFormedProto(value):
		writeProtoMessage(sb, value, depth+1)
	}
}

func isToken(value []byte) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	for _, c := range value {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' && c != '-' {
			return false
		}
	}
	return true
}

func isPrintable(value []byte) bool {
	if !utf8.Valid(value) {
		return false
	}
	for _, r := range string(value) {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func isWellFormedProto(value []byte) bool {
	for len(value) > 0 {
		num, typ, n := protowire.ConsumeTag(value)
		if n < 0 || num > maxProtoFieldNum {
			return false
		}
		value = value[n:]
		switch typ {
		case protowire.VarintType, protowire.Fixed32Type, protowire.Fixed64Type, protowire.BytesType:
			n = protowire.ConsumeFieldValue(num, typ, value)
		default:
			return false
		}
		if n < 0 {
			return false
		}
		value = value[n:]
	}
	return true
}
