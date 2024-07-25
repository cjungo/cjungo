package cjungo

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type HttpServerConf struct {
	Host           *string
	Port           *uint16
	ReadTimeout    *time.Duration
	WriteTimeout   *time.Duration
	MaxHeaderBytes *int
	IsDumpBody     bool
	IsSwag         bool
}

type NewHttpServerDi struct {
	dig.In
	Conf    *HttpServerConf `optional:"true"`
	Handler http.Handler
	Logger  *zerolog.Logger
}

func NewHttpServer(di NewHttpServerDi) *http.Server {
	defaultHost := "127.0.0.1"
	defaultPort := uint16(12345)
	defaultReadTimeout := 10 * time.Second
	defaultWriteTimeout := 10 * time.Second
	defaultMaxHeaderBytes := 1000000
	if di.Conf == nil {
		di.Conf = &HttpServerConf{
			Host:           &defaultHost,
			Port:           &defaultPort,
			ReadTimeout:    &defaultReadTimeout,
			WriteTimeout:   &defaultWriteTimeout,
			MaxHeaderBytes: &defaultMaxHeaderBytes,
			IsDumpBody:     true,
			IsSwag:         true,
		}
		di.Logger.Info().Str("action", "服务器使用默认配置").Msg("[HTTP]")
	} else {
		di.Logger.Info().Str("action", "服务器加载配置").Msg("[HTTP]")
	}
	host := GetOrDefault(di.Conf.Host, defaultHost)
	port := GetOrDefault(di.Conf.Port, defaultPort)
	address := fmt.Sprintf("%s:%d", host, port)
	// TODO 改用中间件
	// readTimeout := GetOrDefault(di.Conf.ReadTimeout, defaultReadTimeout)
	// writeTimeout := GetOrDefault(di.Conf.WriteTimeout, defaultWriteTimeout)
	maxHeaderBytes := GetOrDefault(di.Conf.MaxHeaderBytes, defaultMaxHeaderBytes)

	// 输出服务器信息
	if e := di.Handler.(*echo.Echo); e != nil {
		for i, r := range e.Routes() {
			di.Logger.Info().
				Str("action", "启用路由").
				Int("index", i).
				Str("name", r.Name).
				Str("path", r.Path).
				Str("method", r.Method).
				Msg("[HTTP]")
		}
	}
	di.Logger.Info().Str("action", "服务器监听").Str("address", address).Msg("[HTTP]")

	return &http.Server{
		Addr:    address,
		Handler: di.Handler,
		// TODO 改用中间件
		// ReadTimeout:    readTimeout,
		// WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

func LoadHttpServerConfFromEnv(logger *zerolog.Logger) (*HttpServerConf, error) {
	logger.Info().Str("action", "通过环境变量配置服务器").Msg("[HTTP]")
	conf := &HttpServerConf{
		IsDumpBody: true,
		IsSwag:     true,
	}
	host := os.Getenv("CJUNGO_HTTP_HOST")
	if len(host) > 0 {
		conf.Host = &host
	}

	if err := GetEnvInt("CJUNGO_HTTP_PORT", func(v uint16) {
		conf.Port = &v
	}); err != nil {
		return nil, err
	}
	if err := GetEnvDuration("CJUNGO_HTTP_READ_TIMEOUT", func(v time.Duration) {
		conf.ReadTimeout = &v
	}); err != nil {
		return nil, err
	}
	if err := GetEnvDuration("CJUNGO_HTTP_WRITE_TIMEOUT", func(v time.Duration) {
		conf.WriteTimeout = &v
	}); err != nil {
		return nil, err
	}
	if err := GetEnvInt("CJUNGO_HTTP_MAX_HEADER_BYTES", func(v int) {
		conf.MaxHeaderBytes = &v
	}); err != nil {
		return nil, err
	}

	if err := GetEnvBool("CJUNGO_HTTP_IS_DUMP_BODY", func(v bool) {
		conf.IsDumpBody = v
	}); err != nil {
		return nil, err
	}

	if err := GetEnvBool("CJUNGO_HTTP_IS_SWAG", func(v bool) {
		conf.IsSwag = v
	}); err != nil {
		return nil, err
	}

	return conf, nil
}
