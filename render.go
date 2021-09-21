package ensweb

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (s *Server) RenderJSON(req *Request, model interface{}, status int) *Result {

	if s.debugMode {
		enableCors(&req.w)
	}

	req.w.Header().Set("Content-Type", "application/json")

	res := &Result{
		Status: status,
		Done:   true,
	}

	if model == nil {
		res.Status = http.StatusNoContent
		req.w.WriteHeader(http.StatusNoContent)
	} else {
		req.w.WriteHeader(status)
		enc := json.NewEncoder(req.w)
		enc.Encode(model)
	}
	return res
}

func (s *Server) RenderJSONError(req *Request, status int, errMsg string, logMsg string, args ...interface{}) *Result {
	if logMsg != "" {
		s.log.Error(logMsg, args...)
	}
	model := ErrMessage{
		Error: errMsg,
	}
	return s.RenderJSON(req, model, status)
}

func (s *Server) RenderTemplate(req *Request, renderPath string, model interface{}, status int) *Result {
	templateFile := s.rootPath + renderPath + ".html"
	fmt.Printf("File : %s\n", templateFile)
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		return s.RenderJSON(req, nil, http.StatusNotFound)
	}
	fmt.Printf("File : %s\n", templateFile)
	err = t.Execute(req.w, model)
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		return s.RenderJSON(req, nil, http.StatusInternalServerError)
	}
	fmt.Printf("File : %s\n", templateFile)
	res := &Result{
		Status: status,
		Done:   true,
	}
	return res
}
