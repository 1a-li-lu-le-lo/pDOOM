// Copyright NU Cybernetics. p(DOOM) — research prototype.

package schema

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// CanonicalJSON renders v as canonical JSON: object keys sorted, no
// insignificant whitespace, HTML characters left unescaped, numbers kept as
// written (integers stay integers). The same bytes are produced for equal
// values regardless of struct field order, so the output is suitable for
// hashing and byte-for-byte reproducibility checks.
func CanonicalJSON(v any) ([]byte, error) {
	tree, err := toTree(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := writeCanonical(&buf, tree, "", ""); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CanonicalJSONIndent is CanonicalJSON with two-space indentation and a trailing
// newline; it is the on-disk format of release and candidate files. Lines never
// carry trailing spaces.
func CanonicalJSONIndent(v any) ([]byte, error) {
	tree, err := toTree(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := writeCanonical(&buf, tree, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// SHA256Hex returns the lowercase hex SHA-256 of b.
func SHA256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// toTree round-trips v through encoding/json into generic values while keeping
// number literals verbatim (json.Number).
func toTree(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var tree any
	if err := dec.Decode(&tree); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("canonical json: trailing data")
	}
	return tree, nil
}

func writeCanonical(buf *bytes.Buffer, v any, prefix, indent string) error {
	switch t := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case json.Number:
		buf.WriteString(t.String())
	case string:
		enc, err := encodeString(t)
		if err != nil {
			return err
		}
		buf.Write(enc)
	case []any:
		if len(t) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteByte('[')
		inner := prefix + indent
		for i, el := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			if indent != "" {
				buf.WriteByte('\n')
				buf.WriteString(inner)
			}
			if err := writeCanonical(buf, el, inner, indent); err != nil {
				return err
			}
		}
		if indent != "" {
			buf.WriteByte('\n')
			buf.WriteString(prefix)
		}
		buf.WriteByte(']')
	case map[string]any:
		if len(t) == 0 {
			buf.WriteString("{}")
			return nil
		}
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		inner := prefix + indent
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if indent != "" {
				buf.WriteByte('\n')
				buf.WriteString(inner)
			}
			enc, err := encodeString(k)
			if err != nil {
				return err
			}
			buf.Write(enc)
			buf.WriteByte(':')
			if indent != "" {
				buf.WriteByte(' ')
			}
			if err := writeCanonical(buf, t[k], inner, indent); err != nil {
				return err
			}
		}
		if indent != "" {
			buf.WriteByte('\n')
			buf.WriteString(prefix)
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("canonical json: unsupported value %T", v)
	}
	return nil
}

func encodeString(s string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return []byte(strings.TrimRight(buf.String(), "\n")), nil
}
