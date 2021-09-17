package ensweb

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/EnsurityTechnologies/helper/jsonutil"
	"github.com/EnsurityTechnologies/wraperr"
)

// bufferedReader can be used to replace a request body with a buffered
// version. The Close method invokes the original Closer.
type bufferedReader struct {
	*bufio.Reader
	rOrig io.ReadCloser
}

func newBufferedReader(r io.ReadCloser) *bufferedReader {
	return &bufferedReader{
		Reader: bufio.NewReader(r),
		rOrig:  r,
	}
}

func (b *bufferedReader) Close() error {
	return b.rOrig.Close()
}

func parseQuery(values url.Values) map[string]interface{} {
	data := map[string]interface{}{}
	for k, v := range values {
		// Skip the help key as this is a reserved parameter
		if k == "help" {
			continue
		}

		switch {
		case len(v) == 0:
		case len(v) == 1:
			data[k] = v[0]
		default:
			data[k] = v
		}
	}

	if len(data) > 0 {
		return data
	}
	return nil
}

// isForm tries to determine whether the request should be
// processed as a form or as JSON.
//
// Virtually all existing use cases have assumed processing as JSON,
// and there has not been a Content-Type requirement in the API. In order to
// maintain backwards compatibility, this will err on the side of JSON.
// The request will be considered a form only if:
//
//   1. The content type is "application/x-www-form-urlencoded"
//   2. The start of the request doesn't look like JSON. For this test we
//      we expect the body to begin with { or [, ignoring leading whitespace.
func isForm(head []byte, contentType string) bool {
	contentType, _, err := mime.ParseMediaType(contentType)

	if err != nil || contentType != "application/x-www-form-urlencoded" {
		return false
	}

	// Look for the start of JSON or not-JSON, skipping any insignificant
	// whitespace (per https://tools.ietf.org/html/rfc7159#section-2).
	for _, c := range head {
		switch c {
		case ' ', '\t', '\n', '\r':
			continue
		case '[', '{': // JSON
			return false
		default: // not JSON
			return true
		}
	}

	return true
}

func parseJSONRequest(secondary bool, r *http.Request, w http.ResponseWriter, out interface{}) (io.ReadCloser, error) {
	reader := r.Body
	ctx := r.Context()
	maxRequestSize := ctx.Value("max_request_size")
	if maxRequestSize != nil {
		max, ok := maxRequestSize.(int64)
		if !ok {
			return nil, errors.New("could not parse max_request_size from request context")
		}
		if max > 0 {
			reader = http.MaxBytesReader(w, r.Body, max)
		}
	}

	var origBody io.ReadWriter
	if secondary {
		// Since we're checking PerfStandby here we key on origBody being nil
		// or not later, so we need to always allocate so it's non-nil
		origBody = new(bytes.Buffer)
		reader = ioutil.NopCloser(io.TeeReader(reader, origBody))
	}

	err := jsonutil.DecodeJSONFromReader(reader, out)
	if err != nil && err != io.EOF {
		return nil, wraperr.Wrapf(err, "failed to parse JSON input")
	}
	if origBody != nil {
		return ioutil.NopCloser(origBody), err
	}
	return nil, err
}

// parseFormRequest parses values from a form POST.
//
// A nil map will be returned if the format is empty or invalid.
func parseFormRequest(r *http.Request) (map[string]interface{}, error) {
	maxRequestSize := r.Context().Value("max_request_size")
	if maxRequestSize != nil {
		max, ok := maxRequestSize.(int64)
		if !ok {
			return nil, errors.New("could not parse max_request_size from request context")
		}
		if max > 0 {
			r.Body = ioutil.NopCloser(io.LimitReader(r.Body, max))
		}
	}
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if len(r.PostForm) != 0 {
		data = make(map[string]interface{}, len(r.PostForm))
		for k, v := range r.PostForm {
			switch len(v) {
			case 0:
			case 1:
				data[k] = v[0]
			default:
				// Almost anywhere taking in a string list can take in comma
				// separated values, and really this is super niche anyways
				data[k] = strings.Join(v, ",")
			}
		}
	}

	return data, nil
}

func basicHandleFunc(s *Server, hf HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := basicRequestFunc(s, w, r)

		res := hf(req)
		if res != nil {

		}

	})
}

func (s *Server) ParseJSON(req *Request, model interface{}) error {
	_, err := parseJSONRequest(false, req.r, req.w, model)
	return err
}
