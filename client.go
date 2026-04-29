package http

import (
	"bufio"
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

const defaultTimeout = 30 * time.Second

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

	timeout := c.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	port := req.URL.Port()
	if port == "" {
		switch req.URL.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}

	hostPort := net.JoinHostPort(req.URL.Hostname(), port)
	if hostPort == ":" {
		return nil, fmt.Errorf("invalid host or port")
	}

	tcpConn, err := net.DialTimeout("tcp", hostPort, timeout)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %w", hostPort, err)
	}

	defer func() { _ = tcpConn.Close() }()

	if err := tcpConn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	// req to raw HTTP request bytes
	rawRequest := req.Raw()

	// Write the request bytes
	_, err = tcpConn.Write(rawRequest)
	if err != nil {
		return nil, fmt.Errorf("error writing request: %w", err)
	}

	resp, err := c.readResponse(tcpConn)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return resp, nil
}

func (c *Client) readResponse(conn net.Conn) (*Response, error) {
	reader := bufio.NewReader(conn)

	// Read status line
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading status line: %w", err)
	}

	statusLine = strings.TrimRight(statusLine, "\r\n")

	statusParts := strings.SplitN(statusLine, " ", 3)
	if len(statusParts) < 2 {
		return nil, fmt.Errorf("invalid status line: %s", statusLine)
	}
	
	statusCode, _ := strconv.Atoi(statusParts[1])

	// Read headers
	headers := make(Header)
	var contentLength int64 = -1
	var chunked bool

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading header: %w", err)
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		ck := canonicalKey(key)
		headers[ck] = append(headers[ck], val)
		if ck == "Content-Length" {
			contentLength, _ = strconv.ParseInt(val, 10, 64)
		}
		if ck == "Transfer-Encoding" && strings.EqualFold(val, "chunked") {
			chunked = true
		}
	}

	// Read body
	var body []byte
	if chunked {
		body, err = c.readChunkedBody(reader)
		if err != nil {
			return nil, fmt.Errorf("error reading chunked body: %w", err)
		}
	} else if contentLength >= 0 {
		body = make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
	} else {
		// Read until EOF or connection close
		var buf bytes.Buffer
		_, err = io.Copy(&buf, reader)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
		body = buf.Bytes()
	}

	return &Response{
		Status:     statusLine,
		StatusCode: statusCode,
		Header:     headers,
		Body:       body,
	}, nil
}

func (c *Client) readChunkedBody(reader *bufio.Reader) ([]byte, error) {
	var buf bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if semicolon := strings.Index(line, ";"); semicolon != -1 {
			line = line[:semicolon]
		}
		line = strings.TrimSpace(line)
		chunkSize, err := strconv.ParseInt(line, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chunk size: %s", line)
		}
		if chunkSize < 0 {
			return nil, fmt.Errorf("negative chunk size")
		}
		if chunkSize > 1<<20 {
			return nil, fmt.Errorf("chunk size too large")
		}
		if chunkSize == 0 {
			// Read trailing headers (if any) and final CRLF
			for {
				trailerLine, err := reader.ReadString('\n')
				if err != nil {
					return nil, err
				}
				trailerLine = strings.TrimRight(trailerLine, "\r\n")
				if trailerLine == "" {
					break
				}
			}
			break
		}
		if _, err := io.CopyN(&buf, reader, chunkSize); err != nil {
			return nil, err
		}
		// Read trailing CRLF
		crLf := make([]byte, 2)
		if _, err := io.ReadFull(reader, crLf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (r *Request) Raw() []byte {
	var buf bytes.Buffer

	// Auto-set Content-Length if body exists and header is not set
	if len(r.Body) > 0 && r.Header.Get("Content-Length") == "" {
		r.Header.Set("Content-Length", strconv.Itoa(len(r.Body)))
	}

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
		if !validHeaderFieldName(k) {
			continue
		}
		for _, val := range v {
			if !validHeaderFieldValue(val) {
				continue
			}
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, val))
		}
	}

	// End headers with an empty line
	buf.WriteString("\r\n")

	// Write body if exists
	if r.Body != nil {
		buf.Write(r.Body)
	}

	return buf.Bytes()
}

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
			headers[canonicalKey(key)] = append(headers[canonicalKey(key)], val)
		}
	}

	// Handle chunked encoding
	if te, ok := headers["Transfer-Encoding"]; ok && len(te) > 0 && strings.EqualFold(te[0], "chunked") {
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

func (r *Response) Unmarshal(v any) error {
	return json.Unmarshal(r.Body, v)
}
