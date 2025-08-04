package http_go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseResponse(t *testing.T) {
	raw := []byte(`HTTP/1.1 200 OK
Date: Mon, 27 Jul 2009 12:28:53 GMT
Server: Apache
Last-Modified: Wed, 22 Jul 2009 19:15:56 GMT
ETag: "34aa387-d-1568eb00"
Accept-Ranges: bytes
Content-Length: 51
Vary: Accept-Encoding
Content-Type: text/plain

Hello world, this is the body content from server.`)

	t.Run("parse response", func(t *testing.T) {
		resp, err := parseResponse(raw)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 51, resp.ContentLength)
		assert.Equal(t, "Hello world, this is the body content from server.", string(resp.Body))
	})

}
