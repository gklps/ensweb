package main

import (
	"net/http"
	"time"

	"github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/ensweb"
	"github.com/EnsurityTechnologies/logger"
	"github.com/EnsurityTechnologies/uuid"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
)

// route declaration
const (
	LoginRoute    string = "/api/login"
	LogoutRoute   string = "/api/logout"
	RegisterRoute string = "/api/register"
	HomeRoute     string = "/api/home"
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
	if err != nil {
		return nil, err
	}
	// logOptions := &logger.LoggerOptions{
	// 	Name:  "Audit",
	// 	Color: logger.AutoColor,
	// }

	auditLog := log.Named("audit")
	m, err := NewModel(s.GetDB(), log)
	s.m = m
	s.log = log.Named("exampleserver")
	s.EnableSWagger("ENSWEB Server Example", "This is the exxample server for ensweb framework", "1.0")
	s.SetAuditLog(auditLog)
	s.CreateSessionStore("token-store", "HaiHello", sessions.Options{Path: "/api", HttpOnly: true})
	s.SetDebugMode()
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
	s.AddRoute(LoginRoute, "POST", s.Login)
	s.AddRoute(LogoutRoute, "POST", s.SessionAuthHandle(&Token{}, "token-store", "token", s.Logout, nil))
	s.AddRoute(RegisterRoute, "POST", s.Register)
	s.AddRoute(HomeRoute, "GET", s.SessionAuthHandle(&Token{}, "token-store", "token", s.LoginSession, nil))
	s.SetStatic("./ui/build/")
}

func (s *Server) Index(req *ensweb.Request) *ensweb.Result {
	return s.RenderTemplate(req, "index", nil, http.StatusOK)
}

// ShowAccount godoc
// @Summary      Login into account
// @Description  login in the dashboard
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        login   body      Request  true  "Login Credential"
// @Success      200  {object}  Response
// @Failure      400 {object}  ensweb.ErrMessage
// @Failure      401 {object}  ensweb.ErrMessage
// @Router       /api/login [post]
func (s *Server) Login(req *ensweb.Request) *ensweb.Result {
	var request Request
	err := s.ParseJSON(req, &request)

	if err != nil {
		return s.RenderJSONError(req, http.StatusBadRequest, "Invalid input", "Invalid input")
	}

	user := s.m.GetUser(uuid.Nil, request.UserName)
	if user == nil {
		return s.RenderJSONError(req, http.StatusBadRequest, "User not found", "User not found", "UserName", request.UserName)
	}
	if user.Password != request.Password {
		return s.RenderJSONError(req, http.StatusUnauthorized, "Password mismatch", "Password mismatch", "UserName", request.UserName)
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

	s.SetSessionCookies(req, "token-store", "token", token)

	return s.RenderJSON(req, response, http.StatusOK)
}

// ShowAccount godoc
// @Summary      Logout from the session
// @Description  Logout from the session
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200
// @Failure      507 {object}  ensweb.ErrMessage
// @Router       /api/logout [post]
func (s *Server) Logout(req *ensweb.Request) *ensweb.Result {

	err := s.EmptySessionCookies(req, "token-store")

	if err != nil {
		return s.RenderJSONError(req, http.StatusInsufficientStorage, "Failed clear", "Failed to clear")
	}

	return s.RenderJSON(req, nil, http.StatusOK)

}

// ShowAccount godoc
// @Summary      Register new user account
// @Description  Register new account on the dashboard
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        Request   body      Request  true  "User email & Password"
// @Success      200  {object}  Response
// @Failure      400 {object}  ensweb.ErrMessage
// @Failure      401 {object}  ensweb.ErrMessage
// @Router       /api/register [post]
func (s *Server) Register(req *ensweb.Request) *ensweb.Result {
	var request Request
	isForm, err := s.IsFORM(req)
	if err != nil {
		return s.RenderJSONError(req, http.StatusBadRequest, "Invalid input", "Invalid input")
	}
	if isForm {
		formData, err := s.ParseFORM(req)
		if err != nil {
			return s.RenderJSONError(req, http.StatusBadRequest, "Invalid input", "Invalid input")
		}
		request.UserName = formData["email"].(string)
		request.Password = formData["password"].(string)
	} else {
		err := s.ParseJSON(req, &request)

		if err != nil {
			return s.RenderJSONError(req, http.StatusBadRequest, "Invalid input", "Invalid input")
		}
	}

	user := s.m.GetUser(uuid.Nil, request.UserName)
	if user != nil {
		return s.RenderJSONError(req, http.StatusBadRequest, "User already exist", "User not found", "UserName", request.UserName)
	}
	user = &User{
		UserName: request.UserName,
		Password: request.Password,
	}

	err = s.m.CreateUser(user)
	if err != nil {
		return s.RenderJSONError(req, http.StatusInternalServerError, "User creation failed", "User creation failed", "UserName", request.UserName)
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

	s.SetSessionCookies(req, "token-store", "token", token)

	return s.RenderJSON(req, response, http.StatusOK)
}

// ShowAccount godoc
// @Summary      Login Session
// @Description  Login session in the dashbaord
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Failure      401 {object}  ensweb.ErrMessage
// @Router       /api/home [get]
func (s *Server) LoginSession(req *ensweb.Request) *ensweb.Result {
	if !req.ClientToken.Verified {
		return s.RenderJSONError(req, http.StatusUnauthorized, "Invalid token", "Invalid token")
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
