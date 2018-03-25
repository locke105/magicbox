package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ResponseRecorder wraps a ResponseWriter and
// captures the statusCode and body written.
// headers are already captured in the ResponseWriter,
// and can be accessed on ResponseWriter.Header()
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// WriteHeader captures the HTTP response status code
// when the Handler calls WriteHeader
func (rr *ResponseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the body of the HTTP response when
// the ReponseWriter calls it
func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	rr.body = append(rr.body, b...)
	return rr.ResponseWriter.Write(b)
}

// String returns a string representation of ResponseRecorder
func (rr ResponseRecorder) String() string {
	return fmt.Sprintf(
		"&{statusCode:%d headers:%s body:%s",
		rr.statusCode, rr.ResponseWriter.Header(), rr.body)
}

func (rr ResponseRecorder) Map() map[string]interface{} {
	return map[string]interface{}{
		"statusCode": rr.statusCode,
		"headers":    rr.Header(),
		"body":       fmt.Sprintf("%s", rr.body),
	}
}

type jsonableRequest struct {
	*http.Request
}

func (r jsonableRequest) Map() map[string]interface{} {
	// have to copy the body into a buffer and build a new
	// reader and set it up on the Request struct, otherwise
	// we end up leaving Request.Body at EOF
	bodyBuf, _ := ioutil.ReadAll(r.Body)
	reader := ioutil.NopCloser(bytes.NewBuffer(bodyBuf))
	r.Request.Body = reader

	return map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL,
		"headers": r.Header,
		"body":    bodyBuf,
	}
}

type Jsonable interface {
	Map() map[string]interface{}
}

func prettyJson(rr Jsonable) []byte {
	return _toJson(rr, true)
}

func toJson(rr Jsonable) []byte {
	return _toJson(rr, false)
}

func _toJson(rr Jsonable, pretty bool) []byte {
	var json_str []byte
	var err error

	if pretty {
		json_str, err = json.MarshalIndent(rr.Map(), "", "    ")
	} else {
		json_str, err = json.Marshal(rr.Map())
	}

	if err != nil {
		log.Printf("%+v\n", err)
	}
	return json_str
}

// RecordingProxy is a ReverseProxy that logs all the requests and responses
// that it handles
type RecordingProxy struct {
	httputil.ReverseProxy
}

func (p *RecordingProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Proxying request: %s\n", toJson(jsonableRequest{r}))
	rr := &ResponseRecorder{w, 200, []byte{}}

	p.ReverseProxy.ServeHTTP(rr, r)
	fmt.Printf("Response: %s\n", toJson(rr))
}

func NewRecordingProxy(target *url.URL) *RecordingProxy {
	return &RecordingProxy{*httputil.NewSingleHostReverseProxy(target)}
}
