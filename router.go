package ensweb

import (
	"net/http"
	"strings"

	"github.com/EnsurityTechnologies/ensweb/example/server/docs"
	_ "github.com/EnsurityTechnologies/ensweb/example/server/docs"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// func (s *Server) SetIndex(dirPath string) {
// 	s.mux.Handle("/", s.IndexRoute(dirPath)).Methods("GET")
// }

func (s *Server) AddRoute(path string, method string, hf HandlerFunc) {
	s.mux.Handle(path, basicHandleFunc(s, hf)).Methods(method)
}

func (s *Server) EnableSWagger(title string, description string, version string) {
	url := s.GetServerURL()
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimLeft(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimLeft(url, "https://")
	}
	docs.SwaggerInfo.Title = title
	docs.SwaggerInfo.Description = description
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = url
	docs.SwaggerInfo.BasePath = ""
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	s.mux.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(s.GetServerURL()+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"))).Methods(http.MethodGet)
}

// func (s *Server) SetStatic(dir string) {
// 	s.publicPath = dir
// 	s.mux.PathPrefix("/").Handler(http.FileServer(http.Dir(dir)))
// }

func (s *Server) SetStatic(dir string) {
	s.publicPath = dir
	s.mux.PathPrefix("/").Handler(indexRoute(s, dir))
}

func (s *Server) GetRouteVar(req *Request, key string) string {
	vars := mux.Vars(req.r)
	return vars[key]
}
