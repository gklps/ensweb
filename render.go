package ensweb

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
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

func (s *Server) RenderFile(req *Request, fileName string, attachment bool) *Result {

	if s.debugMode {
		enableCors(&req.w)
	}

	res := &Result{
		Status: http.StatusOK,
		Done:   true,
	}

	if attachment {
		f, err := os.Open(fileName)
		defer f.Close() //Close after function return
		if err != nil {
			//File not found, send 404
			http.Error(req.w, "File not found.", 404)
			res.Status = http.StatusNotFound
			return res
		}

		//File is found, create and send the correct headers

		//Get the Content-Type of the file
		//Create a buffer to store the header of the file in
		FileHeader := make([]byte, 512)
		//Copy the headers into the FileHeader buffer
		f.Read(FileHeader)
		//Get content type of file
		FileContentType := http.DetectContentType(FileHeader)

		//Get the file size
		FileStat, _ := f.Stat()                            //Get info from file
		FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

		//Send the headers
		req.w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		req.w.Header().Set("Content-Type", FileContentType)
		req.w.Header().Set("Content-Length", FileSize)

		//Send the file
		//We read 512 bytes from the file already, so we reset the offset back to 0
		f.Seek(0, 0)
		io.Copy(req.w, f) //'Copy' the file to the client
	} else {
		http.ServeFile(req.w, req.r, fileName)
	}

	return res
}
