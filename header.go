// just simple map functions, nothing fancy
// GET, SET, DELETE, VALUES, LEN, KEYS, CLONE, EXISTS
package http

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = []string{value}
}

func (h Header) Add(key, value string) {
	h[key] = append(h[key], value)
}

func (h Header) Get(key string) string {
	if values, ok := h[key]; ok {
		return values[0]
	}
	return ""
}

func (h Header) Delete(key string) {
	delete(h, key)
}

func (h Header) Values(key string) []string {
	return h[key]
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
		c[k] = v
	}
	return c
}

func (h Header) Exists(key string) bool {
	val, ok := h[key]
	return ok && len(val) > 0
}
