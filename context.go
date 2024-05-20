package cjungo

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type HttpContext interface {
	echo.Context
	GetReqID() string
	RespOk() error
	Resp(any) error
	RespBad(any) error
	RespBadF(string, ...any) error
}

type HttpSimpleContext struct {
	echo.Context
	reqID string
}

func (ctx *HttpSimpleContext) RespOk() error {
	return ctx.JSON(
		http.StatusOK,
		map[string]any{
			"code": 0,
			"data": "Ok",
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

func (ctx *HttpSimpleContext) GetReqID() string {
	return ctx.reqID
}

func ResetContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := uuid.New().String()
		return next(&HttpSimpleContext{
			Context: ctx,
			reqID:   id,
		})
	}
}
