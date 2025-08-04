package http_go

type Method string

const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	HEAD    Method = "HEAD"
	OPTIONS Method = "OPTIONS"
	TRACE   Method = "TRACE"
	CONNECT Method = "CONNECT"
	PATCH   Method = "PATCH"
)

func (m Method) IsValid() bool {
	return m == GET || m == POST || m == PUT || m == DELETE || m == HEAD || m == OPTIONS || m == TRACE || m == CONNECT || m == PATCH
}

func (m Method) IsSafe() bool {
	return m == GET || m == HEAD || m == OPTIONS || m == TRACE || m == CONNECT
}

func (m Method) IsIdempotent() bool {
	return m == GET || m == HEAD || m == PUT || m == DELETE || m == OPTIONS || m == TRACE || m == CONNECT || m == PATCH
}

func (m Method) String() string {
	return string(m)
}
