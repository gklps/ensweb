package ensweb

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/EnsurityTechnologies/adapter"
	"github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/logger"
	"github.com/gorilla/mux"
)

const ServerTimeout = 60 * time.Second

const (
	DefaultTokenHdr  string = "X-Token"
	DefaultRawErrHdr string = "X-Raw"
)

const (
	JSONContentType string = "application/json"
)

type HandlerFunc func(req *Request) *Result

// Server defines server
type Server struct {
	cfg        *config.Config
	serverCfg  *ServerConfig
	s          *http.Server
	mux        *mux.Router
	log        logger.Logger
	auditLog   logger.Logger
	db         *adapter.Adapter
	url        string
	jwtSecret  string
	rootPath   string
	publicPath string
	apiKey     string
	ss         map[string]*SessionStore
	debugMode  bool
}

type ServerConfig struct {
	AuthHeaderName   string
	RawErrHeaderName string
}

type ErrMessage struct {
	Error string `json:"Message"`
}

// NewServer create new server instances
func NewServer(cfg *config.Config, serverCfg *ServerConfig, log logger.Logger) (Server, error) {
	if os.Getenv("ASPNETCORE_PORT") != "" {
		cfg.HostPort = os.Getenv("ASPNETCORE_PORT")
	}
	addr := net.JoinHostPort(cfg.HostAddress, cfg.HostPort)
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  ServerTimeout,
		WriteTimeout: ServerTimeout,
	}
	var serverURL string
	if cfg.Production == "false" {
		serverURL = "http://" + addr
	} else {
		serverURL = "https://" + addr
	}

	db, err := adapter.NewAdapter(cfg)
	if err != nil {
		return Server{}, err
	}

	ts := Server{
		s:          s,
		cfg:        cfg,
		serverCfg:  serverCfg,
		mux:        mux.NewRouter(),
		log:        log.Named("ensweb"),
		db:         db,
		url:        serverURL,
		rootPath:   "views/",
		publicPath: "public/",
		ss:         make(map[string]*SessionStore),
	}

	return ts, nil
}

func (s *Server) SetDebugMode() {
	s.debugMode = true
}

func (s *Server) SetAuditLog(log logger.Logger) {
	s.auditLog = log
}

func (s *Server) SetAPIKey(apiKey string) {
	s.apiKey = apiKey
}

// Start starts the underlying HTTP server
func (s *Server) Start() error {
	// Setup the handler before starting
	s.s.Handler = s.mux
	s.log.Info(fmt.Sprintf("Starting Server at %s", s.s.Addr))
	if s.cfg.Production == "true" {
		go s.s.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile)
		return nil
	} else {
		return s.s.ListenAndServe()
	}
}

// Shutdown attempts to gracefully shutdown the underlying HTTP server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), ServerTimeout)
	defer cancel()
	return s.s.Shutdown(ctx)
}

// GetDB will return DB
func (s *Server) GetDB() *adapter.Adapter {
	return s.db
}
