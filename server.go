package http_go

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultBufSize = 1 << 12
	httpVersion    = "HTTP/1.1"
)

type Server struct {
	Host string
	Port int

	conn map[string]struct{}
	mu   sync.Mutex // for conn, currently only full locking is used, in v2 it will be extended
}

func (s *Server) StartServer() error {
	listener, err := net.Listen(
		"tcp", // for simplicity, I only use tcp
		net.JoinHostPort(s.Host, strconv.Itoa(s.Port)),
	)
	if err != nil {
		return fmt.Errorf("error listening on port %d: %s", s.Port, err)
	}

	log.Printf("Listening on %s:%d", s.Host, s.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %s", err)
		}

		go s.handleConn(conn)
	}
}
func (s *Server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	s.mu.Lock()
	s.conn[conn.RemoteAddr().String()] = struct{}{}
	s.mu.Unlock()

	buf := make([]byte, defaultBufSize)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("error reading from connection: %s", err)
		return
	}

	req, err := parseRequest(buf[:n])
	if err != nil {
		log.Printf("error parsing request: %s", err)
		return
	}
	fmt.Println(">>>>>>>>>>>")

	handler, params := matchRoute(req.Method, req.Path)
	fmt.Println("params: ", params)
	req.pathParams = params
	var resp *Response

	if handler != nil {
		resp = handler(req, params)
	} else {
		resp = &Response{
			StatusCode: 404,
			Headers: Header{
				req.Header.Get("Accept"): {"text/plain"},
			},
			Body: []byte("Not found"),
		}
	}

	writeResponse(conn, resp)
}

// HandlerFunc is a function that takes a request and returns a response
type HandlerFunc func(req *Request, params map[string]string) *Response

var routes = map[Method][]route{}

type route struct {
	pattern string
	method  Method
	handler HandlerFunc
}

func matchRoute(method Method, path string) (HandlerFunc, map[string]string) {
	for _, r := range routes[method] {
		patternParts := strings.Split(r.pattern, "/")
		pathParts := strings.Split(path, "/")

		if len(patternParts) != len(pathParts) {
			continue
		}

		params := make(map[string]string)
		match := true

		for i := range patternParts {
			if strings.HasPrefix(patternParts[i], ":") {
				key := patternParts[i][1:]
				params[key] = pathParts[i]
			} else if patternParts[i] != pathParts[i] {
				match = false
				break
			}
		}

		if match {
			return r.handler, params
		}
	}

	return nil, nil
}

// Route registers a handler for the given method and path
func Route(method Method, path string, handler HandlerFunc) {
	if routes[method] == nil {
		routes[method] = make([]route, 0)
	}

	routes[method] = append(routes[method], route{
		pattern: path,
		method:  method,
		handler: handler,
	})
}

// NewDefaultServer creates a new server with the given host and port
// If host is empty, it will default to "localhost"
// If port is 0, it will default to 8080
func NewDefaultServer(host string, port int) *Server {
	if host == "" {
		host = "localhost"
	}

	if port == 0 {
		port = 8080
	}

	return &Server{
		Host: host,
		Port: port,
		conn: make(map[string]struct{}),
	}
}

// Request Based one: https://datatracker.ietf.org/doc/html/rfc9110#name-connections-clients-and-ser
// GET /hello.txt HTTP/1.1
// User-Agent: curl/7.64.1
// Host: www.example.com
// Accept-Language: en, ru
// Note: the version of the HTTP protocol is 1.1, It might be extended in the future
type Request struct {
	Method  Method
	Path    string
	Query   map[string][]string
	Version string // HTTP/1.1
	Header  Header
	Body    []byte

	pathParams map[string]string
}

var (
	crlf       = []byte("\r\n")
	doubleCRLF = []byte("\r\n\r\n")
	space      = []byte(" ")
)

func parseRequest(data []byte) (*Request, error) {
	if len(data) == 0 {
		return nil, errors.New("empty request")
	}

	req := &Request{
		Header: make(Header),
	}

	// Split header and body
	parts := bytes.SplitN(data, doubleCRLF, 2)
	if len(parts) < 1 {
		return nil, errors.New("invalid http version: " + req.Version)
	}

	headerPart := parts[0]
	var bodyPart []byte
	if len(parts) == 2 {
		bodyPart = parts[1]
	}

	lines := bytes.Split(headerPart, crlf)
	if len(lines) < 1 {
		return nil, errors.New("no request line found")
	}

	// Parse request line
	reqLine := string(lines[0])
	partsReq := strings.SplitN(reqLine, string(space), 3)
	if len(partsReq) != 3 {
		return nil, errors.New("malformed request line")
	}

	method := partsReq[0]

	switch Method(method) {
	case GET, POST, PUT, DELETE:
		req.Method = Method(partsReq[0])
	default:
		return nil, errors.New("unsupported method: " + partsReq[0])
	}

	version := partsReq[2]
	if version != httpVersion {
		return nil, errors.New("unsupported HTTP version: " + version)
	}

	req.Version = version

	// path and query parse
	u, err := url.Parse(partsReq[1])
	if err != nil {
		return nil, err
	}

	req.Path = u.Path
	req.Query = u.Query()

	// Parse headers
	for _, line := range lines[1:] {
		lineStr := string(line)
		if lineStr == "" {
			continue
		}
		colon := strings.Index(lineStr, ":")
		if colon == -1 {
			continue
		}

		key := strings.TrimSpace(lineStr[:colon])
		val := strings.TrimSpace(lineStr[colon+1:])
		req.Header[key] = append(req.Header[key], val)
	}

	if len(bodyPart) > 0 {
		if val, ok := req.Header["content-length"]; ok {
			n, err := strconv.Atoi(val[0])
			if err != nil {
				return nil, errors.New("invalid content-length")
			}
			if n > len(bodyPart) {
				return nil, errors.New("incomplete body")
			}
			req.Body = bodyPart[:n]
		} else if val, ok := req.Header["Transfer-Encoding"]; ok && val[0] == "chunked" {
			decodedBody, err := decodeChunked(bodyPart)
			if err != nil {
				return nil, err
			}
			req.Body = decodedBody

		} else {
			req.Body = bodyPart
		}
	}

	return req, nil
}

func decodeChunked(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	for {
		// Find the next \r\n to get the chunk size line
		i := bytes.Index(data, []byte("\r\n"))
		if i == -1 {
			return nil, errors.New("invalid chunk size line")
		}

		// Parse chunk size (hex)
		chunkSizeHex := string(data[:i])
		chunkSize, err := strconv.ParseInt(strings.TrimSpace(chunkSizeHex), 16, 64)
		if err != nil {
			return nil, errors.New("invalid chunk size: " + chunkSizeHex)
		}

		// Move past chunk size line
		data = data[i+2:]

		if chunkSize == 0 {
			break // last chunk
		}

		if int64(len(data)) < chunkSize+2 {
			return nil, errors.New("incomplete chunk data")
		}

		// Write the chunk data
		buf.Write(data[:chunkSize])

		// Skip chunk data and trailing \r\n
		data = data[chunkSize+2:]
	}

	return buf.Bytes(), nil
}

func (r *Request) PathValue(key string) string {
	if r.pathParams == nil {
		return ""
	}
	return r.pathParams[key]
}

// Response : HTTP/1.1 200 OK
// Date: Mon, 27 Jul 2009 12:28:53 GMT
// Server: Apache
// Last-Modified: Wed, 22 Jul 2009 19:15:56 GMT
// ETag: "34aa387-d-1568eb00"
// Accept-Ranges: bytes
// Content-Length: 51
// Vary: Accept-Encoding
// Content-Type: text/plain
type Response struct {
	StatusCode int
	Status     string
	Headers    Header
	Body       []byte
}

//handleGET(path string, conn net.Conn)	Dispatches GET routes
//writeResponse(conn, status, headers, body)	Forms and sends HTTP response

func writeResponse(conn net.Conn, resp *Response) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s %d %s\r\n", httpVersion, resp.StatusCode, http.StatusText(resp.StatusCode)))
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123) + "\r\n")
	buf.WriteString("Server: Go HTTP Server\r\n")

	for k, v := range resp.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}

	buf.WriteString("Content-Length: " + strconv.Itoa(len(resp.Body)) + "\r\n")

	buf.WriteString("\r\n")
	buf.Write(resp.Body)

	_, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Printf("error writing response: %v\n", err)
	}

}
