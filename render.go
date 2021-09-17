package ensweb

import (
	"encoding/json"
	"net/http"
)

func (s *Server) RenderJSON(req *Request, model interface{}, status int) *Result {
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
