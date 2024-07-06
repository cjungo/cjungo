package cjungo

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type HttpContext interface {
	echo.Context
	GetReqID() string
	GetReqAt() time.Time
	RespOk() error
	Resp(any) error
	RespBad(any) error
	RespBadF(string, ...any) error
}

type HttpSimpleContext struct {
	echo.Context
	reqID string
	reqAt time.Time
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

func (ctx *HttpSimpleContext) RespBad(err error) error {
	return &ApiError{
		Code:     -1,
		Message:  err.Error(),
		HttpCode: http.StatusBadRequest,
		Reason:   err,
	}
}

func (ctx *HttpSimpleContext) RespBadF(format string, data ...any) error {
	return ctx.RespBad(fmt.Errorf(format, data...))
}

func (ctx *HttpSimpleContext) GetReqID() string {
	return ctx.reqID
}

func (ctx *HttpSimpleContext) GetReqAt() time.Time {
	return ctx.reqAt
}

func ResetContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := uuid.New().String()
		now := time.Now()
		return next(&HttpSimpleContext{
			Context: ctx,
			reqID:   id,
			reqAt:   now,
		})
	}
}
