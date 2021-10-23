package ensweb

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func (s *Server) BasicAuthHandle(claims jwt.Claims, hf HandlerFunc, ef HandlerFunc) HandlerFunc {
	return HandlerFunc(func(req *Request) *Result {
		err := s.ValidateJWTToken(req.ClientToken.Token, claims)
		if err != nil {
			if ef != nil {
				return ef(req)
			} else {
				return s.RenderJSONError(req, http.StatusForbidden, err.Error(), err.Error())
			}
		}
		req.ClientToken.Model = claims
		req.ClientToken.Verified = true
		return hf(req)
	})
}

func (s *Server) APIKeyAuthHandle(hf HandlerFunc, ef HandlerFunc) HandlerFunc {
	return HandlerFunc(func(req *Request) *Result {
		if s.apiKey != s.GetReqHeader(req, APIKeyHeader) {
			if ef != nil {
				return ef(req)
			} else {
				return s.RenderJSONError(req, http.StatusForbidden, "API Key is not matched", "API Key is not matched")
			}
		}
		return hf(req)
	})
}
