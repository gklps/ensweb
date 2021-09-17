package ensweb

func (s *Server) AddRoute(path string, hf HandlerFunc) {
	s.mux.Handle(path, basicHandleFunc(s, hf))
}
