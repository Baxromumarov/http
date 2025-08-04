package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	urlPkg "net/url"
	"strconv"
	"strings"
	"time"
)

// Client Request
type Client struct {
	Timeout time.Duration
}

// DefaultClient just empty client
var DefaultClient = &Client{}

func (c *Client) Send(req *Request) (*Response, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered: %v", r)
		}
	}()

	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}

	hostPort := net.JoinHostPort(req.URL.Hostname(), req.URL.Port())
	if hostPort == ":" {
		return nil, fmt.Errorf("invalid host or port")
	}

	tcpConn, err := net.DialTimeout("tcp", hostPort, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %w", hostPort, err)
	}

	defer func() { _ = tcpConn.Close() }()

	if err := tcpConn.SetDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	// req to raw HTTP request bytes
	rawRequest := req.Raw()

	// Write the request bytes
	_, err = tcpConn.Write(rawRequest)
	if err != nil {
		return nil, fmt.Errorf("error writing request: %w", err)
	}

	// Read full response
	var respBuf bytes.Buffer
	tmpBuf := make([]byte, defaultBufSize)

	for {
		n, err := tcpConn.Read(tmpBuf)
		if n > 0 {
			respBuf.Write(tmpBuf[:n])
		}

		if err != nil {
			if err == io.EOF {
				break // finished reading
			}

			// Check if the connection was closed by the server
			if strings.Contains(err.Error(), "connection reset by peer") ||
				strings.Contains(err.Error(), "broken pipe") {
				break // server closed the connection, but we have the response
			}

			return nil, fmt.Errorf("error reading response: %w", err)
		}
	}
	return parseResponse(respBuf.Bytes())
}

func (r *Request) Raw() []byte {
	var buf bytes.Buffer

	// Write request line: METHOD PATH HTTP/1.1\r\n
	buf.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", r.Method, r.URL.RequestURI()))

	//Host header (required for HTTP/1.1)
	host := r.URL.Host
	if host == "" {
		host = r.URL.Hostname()
		if port := r.URL.Port(); port != "" {
			host = net.JoinHostPort(host, port)
		}
	}
	buf.WriteString(fmt.Sprintf("Host: %s\r\n", host))

	//other headers
	for k, v := range r.Header {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}

	// End headers with an empty line
	buf.WriteString("\r\n")

	// Write body if exists
	if r.Body != nil {
		buf.Write(r.Body)
	}

	return buf.Bytes()
}

// HTTP/1.1 200 OK
// Date: Mon, 27 Jul 2009 12:28:53 GMT
// Server: Apache
// Last-Modified: Wed, 22 Jul 2009 19:15:56 GMT
// ETag: "34aa387-d-1568eb00"
// Accept-Ranges: bytes
// Content-Length: 51
// Vary: Accept-Encoding
// Content-Type: text/plain
func parseResponse(raw []byte) (*Response, error) {
	// Split headers and body
	parts := bytes.SplitN(raw, doubleCRLF, 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid response: no header-body separator")
	}

	headerBlock := string(parts[0])
	body := parts[1]

	lines := strings.Split(headerBlock, string(crlf))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	statusLine := lines[0]
	statusCode := 0
	statusParts := strings.SplitN(statusLine, string(space), 3)
	if len(statusParts) >= 2 {
		statusCode, _ = strconv.Atoi(statusParts[1])
	}

	headers := make(Header)
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, string(colon), 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			headers[key] = append(headers[key], val)
		}
	}

	// Handle chunked encoding
	if te, ok := headers["Transfer-Encoding"]; ok && len(te) > 0 && te[0] == "chunked" {
		decodedBody, err := decodeChunked(body)
		if err != nil {
			return nil, fmt.Errorf("error decoding chunked body: %v", err)
		}
		body = decodedBody
	} else if clStr, ok := headers["Content-Length"]; ok {
		// Handle Content-Length
		if len(clStr) == 0 {
			return nil, fmt.Errorf("empty Content-Length")
		}

		cl, err := strconv.Atoi(clStr[0])
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %v", err)
		}

		if len(body) > cl {
			body = body[:cl]
		}
		// If body is shorter than Content-Length, that's okay - it might be truncated
	}

	return &Response{
		Status:     statusLine,
		StatusCode: statusCode,
		Header:     headers,
		Body:       body,
	}, nil
}

func NewRequest(method Method, url string, body []byte) (*Request, error) {
	if url == "" {
		return nil, fmt.Errorf("empty URL")
	}

	u, err := urlPkg.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %s", err)
	}

	return &Request{
		Method:     method,
		URL:        u,
		Body:       body,
		pathParams: make(map[string]string),
		Query:      make(map[string][]string),
		Header:     make(Header),
	}, nil
}

func (r *Response) Unmarshal(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}
