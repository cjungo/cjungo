package cjungo

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type HttpServerConf struct {
	Host           *string
	Port           *uint16
	ReadTimeout    *time.Duration
	WriteTimeout   *time.Duration
	MaxHeaderBytes *int
}

func NewHttpServer(
	conf *HttpServerConf,
	handler http.Handler,
	logger *zerolog.Logger,
) *http.Server {
	host := GetOrDefault(conf.Host, "127.0.0.1")
	port := GetOrDefault(conf.Port, 12345)
	address := fmt.Sprintf("%s:%d", host, port)
	readTimeout := GetOrDefault(conf.ReadTimeout, 10*time.Second)
	writeTimeout := GetOrDefault(conf.WriteTimeout, 10*time.Second)
	maxHeaderBytes := GetOrDefault(conf.MaxHeaderBytes, 1000000)

	// 输出服务器信息
	if e := handler.(*echo.Echo); e != nil {
		for i, r := range e.Routes() {
			logger.Info().
				Int("index", i).
				Str("name", r.Name).
				Str("path", r.Path).
				Str("method", r.Method).
				Msg("route:")
		}
	}
	logger.Info().Str("address", address).Msg("http server listen:")

	return &http.Server{
		Addr:           address,
		Handler:        handler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

func LoadHttpServerConfFromEnv() *HttpServerConf {
	return &HttpServerConf{}
}
