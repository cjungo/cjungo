package cjungo

import (
	"fmt"
	"net/http"
	"time"
)

type HttpServerConf struct {
	Host           *string
	Port           *uint16
	ReadTimeout    *time.Duration
	WriteTimeout   *time.Duration
	MaxHeaderBytes *int
}

func NewHttpServer(conf *HttpServerConf, handler http.Handler) *http.Server {
	host := GetOrDefault(conf.Host, "127.0.0.1")
	port := GetOrDefault(conf.Port, 12345)
	address := fmt.Sprintf("%s:%d", host, port)
	readTimeout := GetOrDefault(conf.ReadTimeout, 10*time.Second)
	writeTimeout := GetOrDefault(conf.WriteTimeout, 10*time.Second)
	maxHeaderBytes := GetOrDefault(conf.MaxHeaderBytes, 1000000)

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
