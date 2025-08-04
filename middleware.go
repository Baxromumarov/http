package http_go

import (
	"context"
	"encoding/base64"
	"log"
	"strings"
	"time"
)

func (s *Server) Use(middleware ...HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware = append(s.middleware, middleware...)
}

func Recover() HandlerFunc {
	return func(req *Request, params map[string]string, next HandlerFunc) *Response {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)
				// Continue to next middleware/handler

			}
		}()

		if next != nil {
			return next(req, params, nil)
		}
		return &Response{
			StatusCode: 500,
			Header:     Header{"Content-Type": {ContentTypeJSON}},
			Body:       []byte("Internal Server Error"),
		}
	}
}
func Logger() HandlerFunc {
	return func(req *Request, params map[string]string, next HandlerFunc) *Response {
		start := time.Now()
		log.Printf("Started %s %s", req.Method, req.Path)

		var resp *Response
		if next != nil {
			resp = next(req, params, nil)
		}

		if resp == nil {
			resp = &Response{
				StatusCode: 200,
				Header:     make(Header),
			}
		}

		log.Printf("Completed %s %s in %v - Status: %d",
			req.Method, req.Path, time.Since(start), resp.StatusCode)

		return resp
	}
}

// CORS adds CORS headers to responses
func CORS() HandlerFunc {
	return func(req *Request, params map[string]string, next HandlerFunc) *Response {
		// Handle preflight request
		if req.Method == "OPTIONS" {
			return &Response{
				StatusCode: 204, // No Content
				Header: Header{
					"Access-Control-Allow-Origin":  {"*"},
					"Access-Control-Allow-Methods": {"GET, POST, PUT, DELETE, OPTIONS"},
					"Access-Control-Allow-Header":  {"Content-Type, Authorization"},
					"Access-Control-Max-Age":       {"86400"}, // 24 hours
				},
			}
		}

		var resp *Response
		if next != nil {
			resp = next(req, params, nil)
		}

		if resp == nil {
			resp = &Response{
				StatusCode: 200,
				Header:     make(Header),
			}
		}

		// Add CORS headers to the response
		resp.Header["Access-Control-Allow-Origin"] = []string{"*"}
		resp.Header["Access-Control-Allow-Methods"] = []string{"GET, POST, PUT, DELETE, OPTIONS"}
		resp.Header["Access-Control-Allow-Header"] = []string{"Content-Type, Authorization"}

		return resp
	}
}

// BasicAuth middleware for HTTP Basic Authentication
func BasicAuth(users map[string]string) HandlerFunc {
	return func(req *Request, params map[string]string, next HandlerFunc) *Response {
		auth := req.Header.Get("Authorization")
		if auth == "" {
			return &Response{
				StatusCode: 401,
				Header:     Header{"WWW-Authenticate": {"Basic realm=\"Restricted\""}},
				Body:       []byte("Unauthorized"),
			}
		}

		// Basic auth format: "Basic base64(username:password)"
		const prefix = "Basic "
		if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
			return &Response{
				StatusCode: 401,
				Header:     Header{"WWW-Authenticate": {"Basic realm=\"Restricted\""}},
				Body:       []byte("Unauthorized"),
			}
		}

		auth = auth[len(prefix):]
		authBytes, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			return &Response{
				StatusCode: 401,
				Header:     Header{"WWW-Authenticate": {"Basic realm=\"Restricted\""}},
				Body:       []byte("Unauthorized"),
			}
		}

		credentials := strings.SplitN(string(authBytes), ":", 2)
		if len(credentials) != 2 {
			return &Response{
				StatusCode: 401,
				Header:     Header{"WWW-Authenticate": {"Basic realm=\"Restricted\""}},
				Body:       []byte("Unauthorized"),
			}
		}

		username, password := credentials[0], credentials[1]
		if pwd, ok := users[username]; !ok || pwd != password {
			return &Response{
				StatusCode: 401,
				Header:     Header{"WWW-Authenticate": {"Basic realm=\"Restricted\""}},
				Body:       []byte("Unauthorized"),
			}
		}

		// Store username in context for later use
		if req.ctx == nil {
			req.ctx = context.Background()
		}
		req.ctx = context.WithValue(req.ctx, "username", username)

		// Authentication successful, continue to next handler
		if next != nil {
			return next(req, params, nil)
		}
		return &Response{StatusCode: 200}
	}
}

// Context returns the request's context
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	return r.ctx
}

// WithContext returns a shallow copy of r with its context changed to ctx
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	return r2
}
