package http_go

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration

	DisableNagle bool // https://en.wikipedia.org/wiki/Nagle%27s_algorithm
	conn         map[string]struct{}
	mu           sync.Mutex // for conn, currently only full locking is used, in v2 it will be extended
	middleware   []MiddlewareFunc
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

func (s *Server) timeoutHandler(conn net.Conn) error {
	if s.ReadTimeout == 0 {
		s.ReadTimeout = 1 * time.Minute
	}

	if s.WriteTimeout == 0 {
		s.WriteTimeout = 1 * time.Minute
	}

	if s.IdleTimeout == 0 {
		s.IdleTimeout = 1 * time.Minute
	}

	if err := conn.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
		return fmt.Errorf("error setting read deadline: %s", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout)); err != nil {
		return fmt.Errorf("error setting write deadline: %s", err)
	}

	if err := conn.SetDeadline(time.Now().Add(s.IdleTimeout)); err != nil {
		return fmt.Errorf("error setting idle deadline: %s", err)
	}

	return nil
}

func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.conn {
		delete(s.conn, k)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	// Setup connection
	if err := s.setupConnection(conn); err != nil {
		log.Printf("error setting up connection: %s", err)
		return
	}

	// Read and parse the HTTP request
	req, err := s.readAndParseRequest(conn)
	if err != nil {
		log.Printf("error reading/parsing request: %s", err)
		return
	}

	// Process the request through middleware and handlers
	resp := s.processRequest(req)

	// Write the response
	writeResponse(conn, resp)
}

// setupConnection configures the connection with timeouts and Nagle's algorithm settings
func (s *Server) setupConnection(conn net.Conn) error {
	// Track connection
	s.mu.Lock()
	s.conn[conn.RemoteAddr().String()] = struct{}{}
	s.mu.Unlock()

	// Configure Nagle's algorithm if disabled
	if s.DisableNagle {
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if err := tcpConn.SetNoDelay(true); err != nil {
				log.Printf("cannot set nodelay: %s", err)
				return err
			}
		} else {
			log.Printf("connection is not a TCP connection")
		}
	}

	// Set timeouts
	return s.timeoutHandler(conn)
}

// readAndParseRequest reads the HTTP request from the connection and parses it
func (s *Server) readAndParseRequest(conn net.Conn) (*Request, error) {
	// Read headers
	headersPart, bodyStart, err := s.readHeaders(conn)
	if err != nil {
		return nil, fmt.Errorf("error reading headers: %w", err)
	}

	// Get content length from headers
	contentLength, err := s.getContentLength(headersPart)
	if err != nil {
		return nil, fmt.Errorf("error getting content length: %w", err)
	}

	// Read the complete body
	fullBody, err := s.readBody(conn, bodyStart, contentLength)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	// Construct and parse the full request
	fullRequest := append(headersPart, fullBody...)
	req, err := parseRequest(fullRequest)
	if err != nil {
		return nil, fmt.Errorf("error parsing request: %w", err)
	}

	return req, nil
}

// readHeaders reads the HTTP headers from the connection
func (s *Server) readHeaders(conn net.Conn) ([]byte, []byte, error) {
	var headerBuf bytes.Buffer
	tmp := make([]byte, 1)

	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			headerBuf.Write(tmp[:n])
			if strings.Contains(headerBuf.String(), "\r\n\r\n") {
				break
			}
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading from connection: %w", err)
		}
	}

	headerData := headerBuf.Bytes()
	headerEnd := bytes.Index(headerData, doubleCRLF)
	if headerEnd == -1 {
		return nil, nil, fmt.Errorf("no header-body separator found")
	}

	headersPart := headerData[:headerEnd+4]
	bodyStart := headerData[headerEnd+4:]

	return headersPart, bodyStart, nil
}

// getContentLength extracts the Content-Length header value
func (s *Server) getContentLength(headersPart []byte) (int, error) {
	// Parse headers to get Content-Length
	tmpReq := &Request{Header: make(Header)}
	lines := bytes.Split(headersPart, crlf)
	if len(lines) < 1 {
		return 0, fmt.Errorf("no request line found")
	}

	// Parse request line
	reqLine := string(lines[0])
	partsReq := strings.SplitN(reqLine, string(space), 3)
	if len(partsReq) != 3 {
		return 0, fmt.Errorf("malformed request line")
	}

	// Parse headers
	for _, line := range lines[1:] {
		lineStr := string(line)
		if lineStr == "" {
			continue
		}
		colon := strings.Index(lineStr, string(colon))
		if colon == -1 {
			continue
		}

		key := strings.TrimSpace(lineStr[:colon])
		val := strings.TrimSpace(lineStr[colon+1:])
		tmpReq.Header[key] = append(tmpReq.Header[key], val)
	}

	// Get Content-Length
	contentLength := 0
	if cl := tmpReq.Header.Get("Content-Length"); cl != "" {
		contentLength, _ = strconv.Atoi(cl)
	}

	return contentLength, nil
}

// readBody reads the complete HTTP body based on Content-Length
func (s *Server) readBody(conn net.Conn, bodyStart []byte, contentLength int) ([]byte, error) {
	if contentLength <= 0 {
		return bodyStart, nil
	}

	remainingBytes := contentLength - len(bodyStart)
	if remainingBytes <= 0 {
		return bodyStart[:contentLength], nil
	}

	remainingBody := make([]byte, remainingBytes)
	if _, err := io.ReadFull(conn, remainingBody); err != nil {
		return nil, fmt.Errorf("error reading remaining body: %w", err)
	}

	return append(bodyStart, remainingBody...), nil
}

// processRequest handles the request through middleware and route matching
func (s *Server) processRequest(req *Request) *Response {
	// Find matching route
	handler, params := matchRoute(req.Method, req.Path)
	req.pathParams = params

	// Create the handler chain with middleware
	handlerChain := s.createHandlerChain(handler)

	// Execute the handler chain
	resp := handlerChain(req)

	// Ensure response has required headers
	s.ensureResponseHeaders(resp)

	return resp
}

// createHandlerChain creates the middleware chain
func (s *Server) createHandlerChain(handler HandlerFunc) HandlerFunc {
	// Start with the route handler
	var handlerChain HandlerFunc = func(req *Request) *Response {
		if handler == nil {
			return &Response{
				StatusCode: 404,
				Header:     Header{"Content-Type": {ContentTypeJSON}},
				Body:       []byte("404 Not found"),
			}
		}
		return handler(req)
	}

	// Apply middleware in reverse order
	for i := len(s.middleware) - 1; i >= 0; i-- {
		mw := s.middleware[i]
		handlerChain = mw(handlerChain)
	}

	return handlerChain
}

// ensureResponseHeaders ensures the response has required headers
func (s *Server) ensureResponseHeaders(resp *Response) {
	if resp.Header == nil {
		resp.Header = make(Header)
	}

	if _, ok := resp.Header["Content-Type"]; !ok && len(resp.Body) > 0 {
		resp.Header["Content-Type"] = []string{ContentTypeJSON}
	}

	// Set Content-Length header
	resp.Header["Content-Length"] = []string{strconv.Itoa(len(resp.Body))}
}

// HandlerFunc is a function that takes a request and returns a response
type HandlerFunc func(req *Request) *Response

// MiddlewareFunc is a function that wraps a handler with additional functionality
type MiddlewareFunc func(HandlerFunc) HandlerFunc

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
			if strings.HasPrefix(patternParts[i], string(colon)) {
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

// Handle registers a handler for the given method and path
func Handle(method Method, path string, handler HandlerFunc) {
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
	Body    []byte // data is always byte array in network, I didn't want to make it some other shit type

	URL  *url.URL
	Host string
	Port int64

	ctx context.Context

	pathParams map[string]string
}

var (
	crlf       = []byte("\r\n")
	doubleCRLF = []byte("\r\n\r\n")
	space      = []byte(" ")
	colon      = []byte(":")
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
		colon := strings.Index(lineStr, string(colon))
		if colon == -1 {
			continue
		}

		key := strings.TrimSpace(lineStr[:colon])
		val := strings.TrimSpace(lineStr[colon+1:])
		req.Header[key] = append(req.Header[key], val)
	}

	if len(bodyPart) > 0 {
		if val := req.Header.Get("Content-Length"); val != "" {
			n, err := strconv.Atoi(val)
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

func (r *Request) QueryValue(key string) []string {
	if r.Query == nil {
		return []string{}
	}
	return r.Query[key]
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
	StatusCode    int
	Status        string
	Header        Header
	Body          []byte
	ContentLength int64
	Request       *Request
}

//handleGET(path string, conn net.Conn)	Dispatches GET routes
//writeResponse(conn, status, headers, body)	Forms and sends HTTP response

func writeResponse(conn net.Conn, resp *Response) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s %d %s\r\n", httpVersion, resp.StatusCode, http.StatusText(resp.StatusCode)))
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123) + "\r\n")
	buf.WriteString("Server: Go HTTP Server\r\n")

	for k, v := range resp.Header {
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

// HTTP method shortcuts
func (s *Server) GET(path string, handler HandlerFunc) {
	Handle(GET, path, handler)
}

func (s *Server) POST(path string, handler HandlerFunc) {
	Handle(POST, path, handler)
}

func (s *Server) PUT(path string, handler HandlerFunc) {
	Handle(PUT, path, handler)
}

func (s *Server) DELETE(path string, handler HandlerFunc) {
	Handle(DELETE, path, handler)
}
