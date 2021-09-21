package ensweb

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/EnsurityTechnologies/uuid"
)

// Operation is an enum that is used to specify the type
// of request being made
type Operation string

type CookiesType map[interface{}]interface{}

type Request struct {
	ID          string
	Method      string
	Path        string
	TimeIn      time.Time
	ClientToken ClientToken
	Connection  *Connection
	Data        map[string]interface{} `json:"data" structs:"data" mapstructure:"data"`
	Model       interface{}
	Headers     http.Header
	r           *http.Request
	w           http.ResponseWriter `json:"-" sentinel:""`
}

type ClientToken struct {
	Token       string
	BearerToken bool
	Verified    bool
	Model       interface{}
}

// Connection represents the connection information for a request.
type Connection struct {
	// RemoteAddr is the network address that sent the request.
	RemoteAddr string `json:"remote_addr"`

	// ConnState is the TLS connection state if applicable.
	ConnState *tls.ConnectionState `sentinel:""`
}

// getConnection is used to format the connection information
func getConnection(r *http.Request) (connection *Connection) {
	var remoteAddr string

	remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteAddr = ""
	}

	connection = &Connection{
		RemoteAddr: remoteAddr,
		ConnState:  r.TLS,
	}
	return
}

func (req *Request) GetHTTPRequest() *http.Request {
	return req.r
}

func basicRequestFunc(s *Server, w http.ResponseWriter, r *http.Request) *Request {

	path := r.URL.Path

	requestId := uuid.New().String()

	req := &Request{
		ID:          requestId,
		Method:      r.Method,
		Path:        path,
		TimeIn:      time.Now(),
		ClientToken: getTokenFromReq(s, r),
		Connection:  getConnection(r),
		Headers:     r.Header,
		r:           r,
		w:           w,
	}

	return req

}

// getTokenFromReq parse headers of the incoming request to extract token if
// present it accepts Authorization Bearer (RFC6750) and configured header.
// Returns true if the token was sourced from a Bearer header.
func getTokenFromReq(s *Server, r *http.Request) ClientToken {
	if s.serverCfg != nil && s.serverCfg.AuthHeaderName != "" {
		if token := r.Header.Get(s.serverCfg.AuthHeaderName); token != "" {
			return ClientToken{Token: token, BearerToken: false}
		}
	}
	if headers, ok := r.Header["Authorization"]; ok {
		// Reference for Authorization header format: https://tools.ietf.org/html/rfc7236#section-3

		// If string does not start by 'Bearer ', it is not one we would use,
		// but might be used by plugins
		for _, v := range headers {
			if !strings.HasPrefix(v, "Bearer ") {
				continue
			}
			return ClientToken{Token: strings.TrimSpace(v[7:]), BearerToken: true}
		}
	}
	return ClientToken{Token: "", BearerToken: false}
}
