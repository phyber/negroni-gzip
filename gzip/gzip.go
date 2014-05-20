package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
)

const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderVary            = "Vary"
)

type gzipResponseWriter struct {
	w *gzip.Writer
	negroni.ResponseWriter
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	if len(grw.Header().Get(HeaderContentType)) == 0 {
		grw.Header().Set(HeaderContentType, http.DetectContentType(b))
	}
	return grw.w.Write(b)
}

type Gzipper struct{}

// Returns our Gzipper that adds gzip compression to all requests
func Gzip() *Gzipper {
	return &Gzipper{}
}

func (g *Gzipper) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !strings.Contains(r.Header.Get(HeaderAcceptEncoding), "gzip") {
		return
	}

	headers := w.Header()
	headers.Set(HeaderContentEncoding, "gzip")
	headers.Set(HeaderVary, HeaderAcceptEncoding)

	gz := gzip.NewWriter(w)
	defer gz.Close()

	gzw := gzipResponseWriter{
		gz,
		w.(negroni.ResponseWriter),
	}

	next(gzw, r)

	// delete content length after we know we have been written to
	gzw.Header().Del("Content-Length")
}
