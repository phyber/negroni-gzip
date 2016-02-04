// Package gzip implements a gzip compression handler middleware for Negroni.
package gzip

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/codegangsta/negroni"
)

// These compression constants are copied from the compress/gzip package.
const (
	encodingGzip = "gzip"

	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerSecWebSocketKey = "Sec-WebSocket-Key"

	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

type Options struct {
	Level               int
	ExcludeContentTypes []string
}

// gzipResponseWriter is the ResponseWriter that negroni.ResponseWriter is
// wrapped in.
type gzipResponseWriter struct {
	w                   *gzip.Writer
	excludeContentTypes []string
	skipCompression     bool
	wroteHeader         bool
	negroni.ResponseWriter
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !grw.wroteHeader {
		// It is not too late to set Content-Type!
		contentType := grw.Header().Get(headerContentType)
		if len(contentType) == 0 {
			contentType = http.DetectContentType(b)
			grw.Header().Set(headerContentType, contentType)
		}

		for _, ct := range grw.excludeContentTypes {
			if contentType == ct {
				grw.skipCompression = true
				break
			}
		}

		if !grw.skipCompression {
			grw.Header().Set(headerContentEncoding, encodingGzip)
			grw.Header().Set(headerVary, headerAcceptEncoding)
		}

		grw.wroteHeader = true
	}

	if grw.skipCompression {
		return grw.ResponseWriter.Write(b) // bypass
	} else {
		return grw.w.Write(b)
	}
}

func (grw *gzipResponseWriter) WriteHeader(s int) {
	contentType := grw.Header().Get(headerContentType)
	for _, ct := range grw.excludeContentTypes {
		if contentType == ct {
			grw.skipCompression = true
			break
		}
	}

	if !grw.skipCompression {
		grw.Header().Set(headerContentEncoding, encodingGzip)
		grw.Header().Set(headerVary, headerAcceptEncoding)
	}

	grw.wroteHeader = true
	grw.ResponseWriter.WriteHeader(s)
}

// handler struct contains the ServeHTTP method
type handler struct {
	pool sync.Pool
	opt  *Options
}

// Gzip returns a handler which will handle the Gzip compression in ServeHTTP.
// Valid values for level are identical to those in the compress/gzip package.
func Gzip(opt *Options) *handler {
	h := &handler{opt: opt}
	h.pool.New = func() interface{} {
		gz, err := gzip.NewWriterLevel(ioutil.Discard, h.opt.Level)
		if err != nil {
			panic(err)
		}
		return gz
	}
	return h
}

// ServeHTTP wraps the http.ResponseWriter with a gzip.Writer.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Skip compression if the client doesn't accept gzip encoding.
	if !strings.Contains(r.Header.Get(headerAcceptEncoding), encodingGzip) {
		next(w, r)
		return
	}

	// Skip compression if client attempt WebSocket connection
	if len(r.Header.Get(headerSecWebSocketKey)) > 0 {
		next(w, r)
		return
	}

	// Skip compression if already compressed
	if w.Header().Get(headerContentEncoding) == encodingGzip {
		next(w, r)
		return
	}

	// Retrieve gzip writer from the pool. Reset it to use the ResponseWriter.
	// This allows us to re-use an already allocated buffer rather than
	// allocating a new buffer for every request.
	// We defer g.pool.Put here so that the gz writer is returned to the
	// pool if any thing after here fails for some reason (functions in
	// next could potentially panic, etc)
	gz := h.pool.Get().(*gzip.Writer)
	defer h.pool.Put(gz)
	gz.Reset(w)

	// Wrap the original http.ResponseWriter with negroni.ResponseWriter
	// and create the gzipResponseWriter.
	nrw := negroni.NewResponseWriter(w)

	grw := gzipResponseWriter{
		w:                   gz,
		excludeContentTypes: h.opt.ExcludeContentTypes,
		ResponseWriter:      nrw,
	}

	// Call the next handler supplying the gzipResponseWriter instead of
	// the original.
	next(&grw, r)

	// Delete the content length after we know we have been written to.
	grw.Header().Del(headerContentLength)

	if !grw.skipCompression {
		gz.Close()
	}
}
