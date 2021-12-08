package ensweb

import "github.com/gorilla/mux"

// func (s *Server) SetIndex(dirPath string) {
// 	s.mux.Handle("/", s.IndexRoute(dirPath)).Methods("GET")
// }

func (s *Server) AddRoute(path string, method string, hf HandlerFunc) {
	s.mux.Handle(path, basicHandleFunc(s, hf)).Methods(method)
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
