package cjungo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HttpContext interface {
	echo.Context
	RespOk() error
	Resp(any) error
	RespBad(any) error
	RespBadF(string, ...any) error
}

type HttpSimpleContext struct {
	echo.Context
}

func (ctx *HttpSimpleContext) RespOk() error {
	return ctx.JSON(
		http.StatusOK,
		map[string]any{
			"code":    0,
			"message": "Ok",
		},
	)
}

func (ctx *HttpSimpleContext) Resp(data any) error {
	return ctx.JSON(
		http.StatusOK,
		map[string]any{
			"code": 0,
			"data": data,
		},
	)
}

func (ctx *HttpSimpleContext) RespBad(data any) error {
	return ctx.JSON(
		http.StatusBadRequest,
		map[string]any{
			"code": -1,
			"data": data,
		},
	)
}

func (ctx *HttpSimpleContext) RespBadF(format string, data ...any) error {
	return ctx.JSON(
		http.StatusBadRequest,
		map[string]any{
			"code": -1,
			"data": fmt.Sprintf(format, data...),
		},
	)
}

func ResetContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return next(&HttpSimpleContext{ctx})
	}
}
