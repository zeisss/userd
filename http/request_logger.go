package http

import (
	"log"
	"net/http"
	"time"
)

type RequestLogger struct {
	Next http.Handler
}

func (l *RequestLogger) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	now := time.Now()

	recorder := ResponseRecorder{ResponseWriter: resp, StatusCode: 200, BytesWritten: 0}

	l.Next.ServeHTTP(&recorder, req)

	taken := time.Since(now)

	log.Printf("%s %s \"%s\" %d %d\n",
		req.Method,
		req.URL,
		req.UserAgent(),
		recorder.StatusCode,
		taken,
	)
}

// --------------------------------------------------------------------------------------------

type ResponseRecorder struct {
	http.ResponseWriter

	StatusCode   int
	BytesWritten int64
}

// Write sums the writes to produce the actual number of bytes written
func (e *ResponseRecorder) Write(b []byte) (int, error) {
	n, err := e.ResponseWriter.Write(b)
	e.BytesWritten += int64(n)
	return n, err
}

// WriteHeader captures the status code and writes through to the wrapper ResponseWriter.
func (e *ResponseRecorder) WriteHeader(code int) {
	e.StatusCode = code
	e.ResponseWriter.WriteHeader(code)
}
