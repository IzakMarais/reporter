package main

import (
	"github.com/gorilla/mux"
	"os"
	"net/http"
	"fmt"
	"io"
	"bytes"
)

type responseWriter struct {
	buf bytes.Buffer
}

func (responseWriter) Header() http.Header {
	return http.Header{}
}

func (responseWriter) WriteHeader(statusCode int) {}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.buf.Write(b)
}

func cmdHandler(router *mux.Router) error {
	fp, err := os.Create(*outputFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	rqStr := "/api/v5/report/%s?apitoken=%s&%s"
	if *apiVersion == "v4" {
		rqStr = "/api/report/%s?apitoken=%s&%s"
	}
	rq, err := http.NewRequest("GET", fmt.Sprintf(rqStr, *dashboard, *apiKey, *timeSpan), nil)
	if err != nil {
		return err
	}
	rw := responseWriter{}
	router.ServeHTTP(&rw, rq)

	_, err = io.Copy(fp, &rw.buf)
	return err
}
