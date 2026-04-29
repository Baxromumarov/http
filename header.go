package http

import (
	"strings"
	"unicode"
)

// Header is a case-insensitive map of HTTP headers.
type Header map[string][]string

// canonicalKey returns the canonical MIME header key for s.
// It is similar to textproto.CanonicalMIMEHeaderKey.
func canonicalKey(s string) string {
	// Fast path: if already canonical, return as-is.
	// A simple check: all letters after a hyphen are uppercase,
	// all other letters are lowercase.
	// We'll just build it.
	return canonicalMIMEHeaderKey(s)
}

func canonicalMIMEHeaderKey(s string) string {
	upper := true
	return strings.Map(
		func(r rune) rune {
			if r == '-' {
				upper = true
				return r
			}

			if upper {
				upper = false
				return unicode.ToUpper(r)
			}

			return unicode.ToLower(r)
		},
		s,
	)
}

func (h Header) Set(key, value string) {
	h[canonicalKey(key)] = []string{value}
}

func (h Header) Add(key, value string) {
	ck := canonicalKey(key)
	h[ck] = append(h[ck], value)
}

func (h Header) Get(key string) string {
	if values, ok := h[canonicalKey(key)]; ok && len(values) > 0 {
		return values[0]
	}

	return ""
}

func (h Header) Delete(key string) {
	delete(h, canonicalKey(key))
}

func (h Header) Values(key string) []string {
	return h[canonicalKey(key)]
}

func (h Header) Len() int {
	return len(h)
}

func (h Header) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

func (h Header) Clone() Header {
	c := make(Header, len(h))
	for k, v := range h {
		c[k] = append([]string(nil), v...)
	}

	return c
}

func (h Header) Exists(key string) bool {
	val, ok := h[canonicalKey(key)]

	return ok && len(val) > 0
}

// validHeaderFieldName reports whether s is a valid HTTP header field name.
func validHeaderFieldName(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if !validHeaderFieldNameRune(r) {
			return false
		}
	}

	return true
}

func validHeaderFieldNameRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '!' ||
		r == '#' ||
		r == '$' ||
		r == '%' ||
		r == '&' ||
		r == '\'' ||
		r == '*' ||
		r == '+' ||
		r == '-' ||
		r == '.' ||
		r == '^' ||
		r == '_' ||
		r == '`' ||
		r == '|' ||
		r == '~'
}

// validHeaderFieldValue reports whether v is a valid HTTP header field value.
// It rejects values containing CR or LF to prevent header injection.
func validHeaderFieldValue(v string) bool {

	for i := 0; i < len(v); i++ {
		c := v[i]

		if c == '\r' || c == '\n' {
			return false
		}
	}

	return true
}
