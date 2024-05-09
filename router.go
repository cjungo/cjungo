package cjungo

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

// TODO 封装 echo.Echo 其他方法

type HttpHandlerFunc func(HttpContext) error

type HttpRouterGroup interface {
	Any(path string, handler HttpHandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
	POST(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Group(prefix string, m ...echo.MiddlewareFunc) (g HttpRouterGroup)
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

func (router *HttpSimpleRouter) Any(path string, h HttpHandlerFunc, m ...echo.MiddlewareFunc) []*echo.Route {
	return router.subject.Any(path, wrapContext(h), m...)
}

func (router *HttpSimpleRouter) Group(prefix string, m ...echo.MiddlewareFunc) (g HttpRouterGroup) {
	return &HttpSimpleGroup{subject: router.subject.Group(prefix, m...)}
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

func (group *HttpSimpleGroup) Group(prefix string, m ...echo.MiddlewareFunc) (g HttpRouterGroup) {
	return &HttpSimpleGroup{subject: group.subject.Group(prefix, m...)}
}

type NewRouterDi struct {
	dig.In
	Logger *zerolog.Logger
	Conf   *HttpServerConf `optional:"true"`
}

func NewRouter(di NewRouterDi) HttpRouter {
	router := echo.New()

	router.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(false),   // e.g. ipv4 start with 127.
		echo.TrustLinkLocal(false),  // e.g. ipv4 start with 169.254
		echo.TrustPrivateNet(false), // e.g. ipv4 start with 10. or 192.168
	)

	// 使用自定义上下文
	router.Use(ResetContext)

	if di.Conf != nil && di.Conf.IsDumpBody {
		router.Use(middleware.BodyDump(func(ctx echo.Context, request, response []byte) {
			di.Logger.Info().
				Str("body", string(request)).
				Msg("请求")

			// TODO 当启用 GZIP 压缩时，信息在日志中是压缩后的数据
			di.Logger.Info().
				Any("body", string(response)).
				Msg("响应")
		}))
	}

	// 异常处理句柄
	router.HTTPErrorHandler = func(err error, c echo.Context) {
		ctx := c.(HttpContext)
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		tip := fmt.Sprintf("%d: %v", code, err)
		di.Logger.Info().Str("error", tip).Msg("error:")
		ctx.RespBad(tip)
	}
	return &HttpSimpleRouter{subject: router}
}
