package traefik_cache

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *ResponseWriter) Write(p []byte) (int, error) {
	rw.body = append(rw.body, p...)
	return rw.ResponseWriter.Write(p)
}

func (rw *ResponseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}
