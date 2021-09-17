package main

import (
	"net/http"
	"time"

	"github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/ensweb"
	"github.com/EnsurityTechnologies/logger"
	"github.com/EnsurityTechnologies/uuid"
	"github.com/dgrijalva/jwt-go"
)

// route declaration
const (
	LoginRoute        string = "/login"
	LoginSessionRoute string = "/loginsession"
)

type Server struct {
	ensweb.Server
	m   *Model
	log logger.Logger
}

// NewServer create new server handle
func NewServer(cfg *config.Config, log logger.Logger) (*Server, error) {
	s := &Server{}
	var err error
	s.Server, err = ensweb.NewServer(cfg, nil, log)
	m, err := NewModel(s.GetDB(), log)
	s.m = m
	s.log = log.Named("exampleserver")
	s.RegisterRoutes()
	return s, err
}

// type HandlerFunc func(context.Context, *Request) (*Response, error)
// type TokenHandleFunc func(token string) (string, bool, interface{}, error)
// type ErrResponseFunc func(req *Request, error string, status int) (*Response

// RegisterRoutes register all routes
func (s *Server) RegisterRoutes() {
	// router := mux.NewRouter()
	// router.HandleFunc("/", s.Index)
	// router.HandleFunc(LoginRoute, s.Login)
	// router.HandleFunc(LoginSessionRoute, s.LoginSession)
	s.AddRoute("/", s.Index)
	s.AddRoute(LoginRoute, s.Login)
	s.AddRoute(LoginSessionRoute, s.BasicAuthHandle(&Token{}, s.LoginSession, nil))
}

func (s *Server) Index(req *ensweb.Request) *ensweb.Result {
	return s.RenderJSONError(req, http.StatusUnauthorized, "Invalid Session", "Invalid Session")
}

func (s *Server) Login(req *ensweb.Request) *ensweb.Result {

	var request Request
	err := s.ParseJSON(req, &request)

	if err != nil {
		return s.RenderJSONError(req, http.StatusBadRequest, "Invalid input", "Invalid input")
	}

	user := s.m.GetUser(uuid.Nil, request.UserName)
	if user == nil {
		return s.RenderJSONError(req, http.StatusForbidden, "User not found", "User not found", "UserName", request.UserName)
	}
	if user.Password != request.Password {
		return s.RenderJSONError(req, http.StatusForbidden, "Password mismatch", "Password mismatch", "UserName", request.UserName)
	}
	expiresAt := time.Now().Add(time.Minute * 1).Unix()

	claims := Token{
		request.UserName,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := s.GenerateJWTToken(claims)

	response := Response{
		Token: token,
	}

	return s.RenderJSON(req, response, http.StatusOK)
}

func (s *Server) LoginSession(req *ensweb.Request) *ensweb.Result {
	if !req.ClientToken.Verified {
		return s.RenderJSONError(req, http.StatusForbidden, "Invalid token", "Invalid token")
	}
	claims := req.ClientToken.Model.(*Token)

	// resp := &ensweb.Response{
	// 	Data: map[string]interface{}{},
	// }
	// resp.Data["Message"] = "Valid User Session : " + claims.UserNameresponse := Response{
	// 	Token: token,
	// }

	response := Response{
		Message: "Valid User Session : " + claims.UserName,
	}

	return s.RenderJSON(req, response, http.StatusOK)
}
