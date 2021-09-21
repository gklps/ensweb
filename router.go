package ensweb

import "net/http"

func (s *Server) AddRoute(path string, method string, hf HandlerFunc) {
	s.mux.Handle(path, basicHandleFunc(s, hf)).Methods(method)
}

func (s *Server) SetStatic(dir string) {
	s.publicPath = dir
	s.mux.PathPrefix("/").Handler(http.FileServer(http.Dir(dir)))
}
