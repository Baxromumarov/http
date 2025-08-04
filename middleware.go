package http

import (
	"context"
	"encoding/base64"
	"log"
	"strings"
	"time"
)

func (s *Server) Use(middleware ...MiddlewareFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware = append(s.middleware, middleware...)
}

func Recover() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) *Response {
			var resp *Response
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic recovered: %v", r)
						resp = &Response{
							StatusCode: 500,
							Header:     Header{"Content-Type": {ContentTypeJSON}},
							Body:       []byte("Internal Server Error"),
						}
					}
				}()

				resp = next(req)
			}()

			// Ensure CORS headers are present even after panic
			if resp != nil && resp.Header == nil {
				resp.Header = make(Header)
			}
			if resp != nil {
				resp.Header["Access-Control-Allow-Origin"] = []string{"*"}
				resp.Header["Access-Control-Allow-Methods"] = []string{"GET, POST, PUT, DELETE, OPTIONS"}
				resp.Header["Access-Control-Allow-Headers"] = []string{"Content-Type, Authorization"}
			}

			return resp
		}
	}
}
func Logger() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) *Response {
			start := time.Now()
			log.Printf("Started %s %s", req.Method, req.Path)

			resp := next(req)

			log.Printf("Completed %s %s in %v - Status: %d",
				req.Method, req.Path, time.Since(start), resp.StatusCode)

			return resp
		}
	}
}

// CORS adds CORS headers to responses
func CORS() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) *Response {
			// Handle preflight request
			if req.Method == "OPTIONS" {
				return &Response{
					StatusCode: 204, // No Content
					Header: Header{
						"Access-Control-Allow-Origin":  {"*"},
						"Access-Control-Allow-Methods": {"GET, POST, PUT, DELETE, OPTIONS"},
						"Access-Control-Allow-Headers": {"Content-Type, Authorization"},
						"Access-Control-Max-Age":       {"86400"}, // 24 hours
					},
				}
			}

			resp := next(req)

			// Add CORS headers to the response
			if resp.Header == nil {
				resp.Header = make(Header)
			}
			resp.Header["Access-Control-Allow-Origin"] = []string{"*"}
			resp.Header["Access-Control-Allow-Methods"] = []string{"GET, POST, PUT, DELETE, OPTIONS"}
			resp.Header["Access-Control-Allow-Headers"] = []string{"Content-Type, Authorization"}

			return resp
		}
	}
}

// BasicAuth middleware for HTTP Basic Authentication
func BasicAuth(users map[string]string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) *Response {
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

			if req.ctx == nil {
				req.ctx = context.Background()
			}
			req.ctx = context.WithValue(req.ctx, "username", username)

			return next(req)
		}
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
