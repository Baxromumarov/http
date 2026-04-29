package http

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// defaultBufSize   = 1 << 12 // 4 KB
	httpVersion      = "HTTP/1.1"
	defaultMaxHeader = 1 << 20 // 1 MB
	defaultMaxBody   = 1 << 22 // 4 MB
)

type Server struct {
	Host string
	Port int

	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	IdleTimeout  time.Duration

	MaxHeaderBytes int
	MaxBodyBytes   int64

	DisableNagle bool // https://en.wikipedia.org/wiki/Nagle%27s_algorithm
	conn         map[string]struct{}
	mu           sync.Mutex // for conn and middleware
	middleware   []MiddlewareFunc

	router *Router
}

type ServerI interface {
	Handle(method Method, path string, handler HandlerFunc)
	GET(path string, handler HandlerFunc)
	POST(path string, handler HandlerFunc)
	PUT(path string, handler HandlerFunc)
	DELETE(path string, handler HandlerFunc)
	HEAD(path string, handler HandlerFunc)
	OPTIONS(path string, handler HandlerFunc)
	PATCH(path string, handler HandlerFunc)
	init()
	StartServer() error
	Close()
}

var _ ServerI = (*Server)(nil) // compile-time check that Server implements ServerI

// Router holds registered routes protected by a RWMutex.
type Router struct {
	mu     sync.RWMutex
	routes map[Method][]route
}

type route struct {
	pattern string
	method  Method
	handler HandlerFunc
}

func (r *Router) Handle(
	method Method,
	path string,
	handler HandlerFunc,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.routes == nil {
		r.routes = make(map[Method][]route)
	}

	r.routes[method] = append(r.routes[method], route{
		pattern: path,
		method:  method,
		handler: handler,
	})
}

func (r *Router) match(
	method Method,
	path string,
) (
	HandlerFunc,
	map[string]string,
) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rt := range r.routes[method] {
		patternParts := strings.Split(rt.pattern, "/")
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
			return rt.handler, params
		}
	}

	return nil, nil
}

func (s *Server) init() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		s.conn = make(map[string]struct{})
	}
	if s.router == nil {
		s.router = &Router{}
	}
	if s.MaxHeaderBytes == 0 {
		s.MaxHeaderBytes = defaultMaxHeader
	}
	if s.MaxBodyBytes == 0 {
		s.MaxBodyBytes = defaultMaxBody
	}
}

func (s *Server) StartServer() error {
	s.init()

	listener, err := net.Listen(
		"tcp", // for simplicity, I only used tcp
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
	readTimeout := s.ReadTimeout
	writeTimeout := s.WriteTimeout
	idleTimeout := s.IdleTimeout

	if readTimeout == 0 {
		readTimeout = defaultTimeout
	}
	if writeTimeout == 0 {
		writeTimeout = defaultTimeout
	}
	if idleTimeout == 0 {
		idleTimeout = defaultTimeout
	}

	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return fmt.Errorf("error setting read deadline: %s", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("error setting write deadline: %s", err)
	}

	if err := conn.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
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
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered in connection handler: %v", r)
			writeErrorResponse(conn, StatusInternalServerError)
		}
		_ = conn.Close()
	}()

	// Setup connection
	if err := s.setupConnection(conn); err != nil {
		log.Printf("error setting up connection: %s", err)
		return
	}

	// Read and parse the HTTP request
	req, err := s.readAndParseRequest(conn)
	if err != nil {
		log.Printf("error reading/parsing request: %s", err)
		if errors.Is(err, ErrRequestEntityTooLarge) {
			writeErrorResponse(conn, StatusRequestEntityTooLarge)
		} else {
			writeErrorResponse(conn, StatusBadRequest)
		}
		return
	}

	// Process the request through middleware and handlers
	resp := s.processRequest(req)

	// Write the response
	writeResponse(conn, resp)
}

var ErrRequestEntityTooLarge = errors.New("request entity too large")

func writeErrorResponse(conn net.Conn, code int) {
	body := []byte(StatusText(code))

	resp := &Response{
		StatusCode: code,
		Header:     Header{"Content-Type": {ContentTypeText}},
		Body:       body,
	}

	writeResponse(conn, resp)
}

// setupConnection configures the connection with timeouts and Nagle's algorithm settings
func (s *Server) setupConnection(conn net.Conn) error {
	s.init()

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
	reader := bufio.NewReader(conn)

	// Read headers
	headersPart, bodyStart, err := s.readHeaders(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading headers: %w", err)
	}

	// Get content length from headers
	contentLength, err := s.getContentLength(headersPart)
	if err != nil {
		return nil, fmt.Errorf("error getting content length: %w", err)
	}

	// Read the complete body
	fullBody, err := s.readBody(reader, bodyStart, contentLength)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	// Check max body size
	if s.MaxBodyBytes > 0 &&
		int64(len(fullBody)) > s.MaxBodyBytes {
		return nil, ErrRequestEntityTooLarge
	}

	// Construct and parse the full request
	fullRequest := append(headersPart, fullBody...)
	req, err := parseRequest(fullRequest)
	if err != nil {
		return nil, fmt.Errorf("error parsing request: %w", err)
	}

	return req, nil
}

// readHeaders reads the HTTP headers from the connection using bufio.
func (s *Server) readHeaders(
	reader *bufio.Reader,
) (
	headersPart []byte,
	bodyStart []byte,
	err error,
) {
	var headerBuf bytes.Buffer

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("error reading from connection: %w", err)
		}

		headerBuf.Write(line)

		if s.MaxHeaderBytes > 0 &&
			headerBuf.Len() > s.MaxHeaderBytes {
			return nil, nil, ErrRequestEntityTooLarge
		}

		if bytes.HasSuffix(headerBuf.Bytes(), doubleCRLF) {
			break
		}
	}

	headerData := headerBuf.Bytes()
	headerEnd := bytes.Index(headerData, doubleCRLF)
	if headerEnd == -1 {
		return nil, nil, fmt.Errorf("no header-body separator found")
	}

	headersPart = headerData[:headerEnd+4]
	bodyStart = headerData[headerEnd+4:]

	return headersPart, bodyStart, nil
}

// getContentLength extracts the Content-Length header value with validation.
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
	var contentLengthSet bool
	var contentLength int
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
		ck := canonicalKey(key)

		if ck == "Content-Length" {
			cl, err := strconv.Atoi(val)
			if err != nil {
				return 0, fmt.Errorf("invalid content-length")
			}

			if cl < 0 {
				return 0, fmt.Errorf("negative content-length")
			}

			if contentLengthSet && cl != contentLength {
				return 0, fmt.Errorf("conflicting content-length values")
			}

			contentLength = cl
			contentLengthSet = true
		}

		tmpReq.Header[ck] = append(tmpReq.Header[ck], val)
	}

	return contentLength, nil
}

// readBody reads the complete HTTP body based on Content-Length
func (s *Server) readBody(
	reader *bufio.Reader,
	bodyStart []byte,
	contentLength int,
) (
	[]byte,
	error,
) {
	if contentLength <= 0 {
		return bodyStart, nil
	}

	if s.MaxBodyBytes > 0 &&
		int64(contentLength) > s.MaxBodyBytes {

		return nil, ErrRequestEntityTooLarge
	}

	remainingBytes := contentLength - len(bodyStart)
	if remainingBytes <= 0 {
		return bodyStart[:contentLength], nil
	}

	remainingBody := make([]byte, remainingBytes)
	if _, err := io.ReadFull(reader, remainingBody); err != nil {
		return nil, fmt.Errorf("error reading remaining body: %w", err)
	}

	return append(bodyStart, remainingBody...), nil
}

// processRequest handles the request through middleware and route matching
func (s *Server) processRequest(req *Request) *Response {
	// Find matching route
	handler, params := s.router.match(req.Method, req.Path)
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
				StatusCode: StatusNotFound,
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

// Handle registers a handler on the default global server.
// For multi-server usage, use server.Handle instead.
func Handle(
	method Method,
	path string,
	handler HandlerFunc,
) {
	defaultServer.Handle(method, path, handler)
}

var defaultServer = &Server{
	router: &Router{},
	conn:   make(map[string]struct{}),
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
		Host:   host,
		Port:   port,
		conn:   make(map[string]struct{}),
		router: &Router{},
	}
}

// Request Based on: https://datatracker.ietf.org/doc/html/rfc9110#name-connections-clients-and-ser
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
		return nil, errors.New("invalid http request")
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

	method := Method(partsReq[0])
	if !method.IsValid() {
		return nil, errors.New("unsupported method: " + partsReq[0])
	}
	req.Method = method

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
		colonIdx := strings.Index(lineStr, string(colon))
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(lineStr[:colonIdx])
		val := strings.TrimSpace(lineStr[colonIdx+1:])
		cKey := canonicalKey(key)

		req.Header[cKey] = append(req.Header[cKey], val)
	}

	// Validate Content-Length regardless of body presence
	if clValues := req.Header.Values("Content-Length"); len(clValues) > 0 {
		var contentLength int
		for i, v := range clValues {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, errors.New("invalid content-length")
			}

			if n < 0 {
				return nil, errors.New("negative content-length")
			}

			if i == 0 {
				contentLength = n
			} else if n != contentLength {
				return nil, errors.New("conflicting content-length values")
			}
		}

		if len(bodyPart) > 0 {
			if contentLength > len(bodyPart) {
				return nil, errors.New("incomplete body")
			}
			req.Body = bodyPart[:contentLength]
		}

	} else if len(bodyPart) > 0 {
		if val, ok := req.Header["Transfer-Encoding"]; ok &&
			len(val) > 0 &&
			strings.EqualFold(val[0], "chunked") {

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
		// Chunk extensions are separated by semicolon
		if semicolon := strings.Index(chunkSizeHex, ";"); semicolon != -1 {
			chunkSizeHex = chunkSizeHex[:semicolon]
		}

		chunkSizeHex = strings.TrimSpace(chunkSizeHex)
		chunkSize, err := strconv.ParseInt(chunkSizeHex, 16, 64)
		if err != nil {
			return nil, errors.New("invalid chunk size: " + chunkSizeHex)
		}

		if chunkSize < 0 {
			return nil, errors.New("negative chunk size")
		}

		if chunkSize > defaultMaxHeader {
			return nil, errors.New("chunk size too large")
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

// writeResponse forms and sends HTTP response
func writeResponse(conn net.Conn, resp *Response) {
	var buf bytes.Buffer

	fmt.Fprintf(
		&buf,
		"%s %d %s\r\n",
		httpVersion,
		resp.StatusCode,
		StatusText(resp.StatusCode),
	)

	fmt.Fprintf(
		&buf,
		"Date: %s\r\n",
		time.Now().Format(time.RFC1123),
	)

	fmt.Fprintf(&buf, "Server: Go HTTP Server\r\n")

	for k, v := range resp.Header {
		if !validHeaderFieldName(k) {
			continue
		}

		for _, val := range v {
			if !validHeaderFieldValue(val) {
				continue
			}

			fmt.Fprintf(&buf, "%s: %s\r\n", k, val)
		}
	}

	fmt.Fprintf(
		&buf,
		"Content-Length: %d\r\n",
		len(resp.Body),
	)

	fmt.Fprint(&buf, "\r\n")
	fmt.Fprint(&buf, string(resp.Body))

	_, err := conn.Write(buf.Bytes())
	if err != nil {
		log.Printf("error writing response: %v\n", err)
	}
}

// Handle registers a handler for the given method and path on this server's router.
func (s *Server) Handle(
	method Method,
	path string,
	handler HandlerFunc,
) {
	s.init()
	s.router.Handle(method, path, handler)
}

// HTTP method shortcuts
func (s *Server) GET(path string, handler HandlerFunc) {
	s.Handle(GET, path, handler)
}

func (s *Server) POST(path string, handler HandlerFunc) {
	s.Handle(POST, path, handler)
}

func (s *Server) PUT(path string, handler HandlerFunc) {
	s.Handle(PUT, path, handler)
}

func (s *Server) DELETE(path string, handler HandlerFunc) {
	s.Handle(DELETE, path, handler)
}

func (s *Server) HEAD(path string, handler HandlerFunc) {
	s.Handle(HEAD, path, handler)
}

func (s *Server) OPTIONS(path string, handler HandlerFunc) {
	s.Handle(OPTIONS, path, handler)
}

func (s *Server) PATCH(path string, handler HandlerFunc) {
	s.Handle(PATCH, path, handler)
}
