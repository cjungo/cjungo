package cjungo

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
)

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	err := responseControllerFlush(w.ResponseWriter)
	if err != nil && errors.Is(err, http.ErrNotSupported) {
		panic(errors.New("response writer flushing is not supported"))
	}
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return responseControllerHijack(w.ResponseWriter)
}

func (w *bodyDumpResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func responseControllerFlush(rw http.ResponseWriter) error {
	return http.NewResponseController(rw).Flush()
}

func responseControllerHijack(rw http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(rw).Hijack()
}

type DumpBodyHandle func(ctx HttpContext, req []byte, resp []byte) error

func NewDumpBodyMiddleware(handle DumpBodyHandle) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ctx, ok := c.(HttpContext); ok {
				// Request
				reqBody := []byte{}
				if c.Request().Body != nil { // Read
					if b, err := io.ReadAll(c.Request().Body); err != nil {
						ctx.Error(err)
						return err
					} else {
						reqBody = b
					}
				}
				c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset

				// Response
				resBody := new(bytes.Buffer)
				mw := io.MultiWriter(c.Response().Writer, resBody)
				writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
				c.Response().Writer = writer

				if err := handle(ctx, reqBody, resBody.Bytes()); err != nil {
					ctx.Error(err)
					return err
				}

			}
			return next(c)
		}
	}
}
