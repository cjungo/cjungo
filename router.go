package cjungo

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/dig"
)

// TODO 封装 echo.Echo 其他方法

type HttpHandlerFunc func(HttpContext) error

type HttpRouterGroup interface {
	Any(path string, handler HttpHandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
	POST(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Group(prefix string, m ...echo.MiddlewareFunc) (g HttpRouterGroup)
	Use(middleware ...echo.MiddlewareFunc)
}

type HttpRouter interface {
	HttpRouterGroup
	GetHandler() http.Handler
	Static(pathPrefix string, fsRoot string) *echo.Route
	StaticFS(pathPrefix string, filesystem fs.FS) *echo.Route
}

type HttpSimpleRouter struct {
	subject *echo.Echo
}

func wrapContext(h HttpHandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(HttpContext)
		return h(ctx)
	}
}

func (router *HttpSimpleRouter) GET(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return router.subject.GET(path, wrapContext(h), m...)
}

func (router *HttpSimpleRouter) PUT(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return router.subject.PUT(path, wrapContext(h), m...)
}

func (router *HttpSimpleRouter) POST(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return router.subject.POST(path, wrapContext(h), m...)
}

func (router *HttpSimpleRouter) DELETE(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return router.subject.DELETE(path, wrapContext(h), m...)
}

func (router *HttpSimpleRouter) Any(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) []*echo.Route {
	return router.subject.Any(path, wrapContext(h), m...)
}

func (router *HttpSimpleRouter) Group(prefix string, m ...echo.MiddlewareFunc) (g HttpRouterGroup) {
	return &HttpSimpleGroup{subject: router.subject.Group(prefix, m...)}
}

func (router *HttpSimpleRouter) Use(middleware ...echo.MiddlewareFunc) {
	router.subject.Use(middleware...)
}

func (router *HttpSimpleRouter) GetHandler() http.Handler {
	return router.subject
}

func (router *HttpSimpleRouter) Static(pathPrefix string, fsRoot string) *echo.Route {
	return router.subject.Static(pathPrefix, fsRoot)
}
func (router *HttpSimpleRouter) StaticFS(pathPrefix string, filesystem fs.FS) *echo.Route {
	return router.subject.StaticFS(pathPrefix, filesystem)
}

type HttpSimpleGroup struct {
	subject *echo.Group
}

func (group *HttpSimpleGroup) Any(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) []*echo.Route {
	return group.subject.Any(path, wrapContext(h), m...)
}

func (group *HttpSimpleGroup) GET(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return group.subject.GET(path, wrapContext(h), m...)
}

func (group *HttpSimpleGroup) PUT(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return group.subject.PUT(path, wrapContext(h), m...)
}

func (group *HttpSimpleGroup) POST(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return group.subject.POST(path, wrapContext(h), m...)
}

func (group *HttpSimpleGroup) DELETE(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return group.subject.DELETE(path, wrapContext(h), m...)
}

func (group *HttpSimpleGroup) Group(prefix string, m ...echo.MiddlewareFunc) (g HttpRouterGroup) {
	return &HttpSimpleGroup{subject: group.subject.Group(prefix, m...)}
}

func (group *HttpSimpleGroup) Use(middleware ...echo.MiddlewareFunc) {
	group.subject.Use(middleware...)
}

type NewRouterDi struct {
	dig.In
	Logger *zerolog.Logger
	Conf   *HttpServerConf `optional:"true"`
}

type RouterLogger struct {
	subject *zerolog.Logger
}

func (logger RouterLogger) Write(p []byte) (n int, err error) {
	logger.subject.Info().RawJSON("echo", p).Msg("[HTTP]")
	return len(p), nil
}

func NewRouter(di NewRouterDi) HttpRouter {
	router := echo.New()

	router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_custom}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.000",
		Output:           &RouterLogger{subject: di.Logger},
	}))

	router.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(false),   // e.g. ipv4 start with 127.
		echo.TrustLinkLocal(false),  // e.g. ipv4 start with 169.254
		echo.TrustPrivateNet(false), // e.g. ipv4 start with 10. or 192.168
	)

	// 使用自定义上下文
	router.Use(ResetContext)

	if di.Conf != nil && di.Conf.IsDumpBody {
		router.Use(NewDumpBodyMiddleware(func(ctx HttpContext, req, resp []byte) error {
			di.Logger.Info().
				Str("body", string(req)).
				Str("action", "打印请求内容").
				Msg("[HTTP]")

			// TODO 当启用 GZIP 压缩时，信息在日志中是压缩后的数据
			di.Logger.Info().
				Any("body", string(resp)).
				Str("action", "打印响应内容").
				Msg("[HTTP]")
			return nil
		}))
	}

	// 错误处理句柄
	router.HTTPErrorHandler = func(err error, ctx echo.Context) {
		var result *ApiError
		if apiError, ok := err.(*ApiError); ok {
			result = apiError
		} else {
			result = &ApiError{
				Code:     -1,
				Message:  err.Error(),
				HttpCode: http.StatusInternalServerError,
				Reason:   err,
			}
			if httpError, ok := err.(*echo.HTTPError); ok {
				result.HttpCode = httpError.Code
			}
		}

		di.Logger.Error().
			Stack().
			Int("code", result.HttpCode).
			Err(err).
			Msg("[HTTP]")

		ctx.JSON(result.HttpCode, result)
	}

	if di.Conf != nil && di.Conf.IsSwag {
		link := fmt.Sprintf("http://%s:%d/swagger/", *di.Conf.Host, *di.Conf.Port)
		router.GET("/swagger/*", echoSwagger.WrapHandler)
		di.Logger.Info().Str("link", link).Msg("[SWAG]")
	}

	return &HttpSimpleRouter{subject: router}
}
