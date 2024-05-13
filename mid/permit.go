package mid

import (
	"sync"

	"github.com/cjungo/cjungo"
	"github.com/elliotchance/pie/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/exp/constraints"
)

type Permission interface {
	constraints.Integer | string
}

type AuthKeyHandle[TP Permission, TS any] func(cjungo.HttpContext) ([]TP, TS, error)

type PermitManager[TP Permission, TS any] struct {
	logger *zerolog.Logger
	tokens sync.Map
	handle AuthKeyHandle[TP, TS]
}

type PermitManagerProvide[TP Permission, TS any] func(
	logger *zerolog.Logger,
) (*PermitManager[TP, TS], error)

func NewPermitManager[TP Permission, TS any](handle AuthKeyHandle[TP, TS]) PermitManagerProvide[TP, TS] {
	return func(logger *zerolog.Logger) (*PermitManager[TP, TS], error) {
		manager := &PermitManager[TP, TS]{
			logger: logger,
			handle: handle,
		}
		return manager, nil
	}
}

func (manager *PermitManager[TP, TS]) GetToken(id string, ts *TS) bool {
	v, ok := manager.tokens.Load(id)
	if ok {
		*ts = v.(TS)
	}
	return ok
}

func (manager *PermitManager[TP, TS]) Permit(permissions ...TP) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.(cjungo.HttpContext)
			manager.logger.Info().Any("reqID", ctx.GetReqID()).Msg("[PermitManager]")
			if owned, ts, err := manager.handle(ctx); err != nil {
				return err
			} else {
				reqID := ctx.GetReqID()
				manager.tokens.Store(reqID, ts)
				manager.logger.Info().Any("store", ts).Str("id", reqID).Msg("[PermitManager]")
				defer func() {
					manager.logger.Info().Any("delete", ts).Str("id", reqID).Msg("[PermitManager]")
					manager.tokens.Delete(reqID)
				}()
				added, _ := pie.Diff(owned, permissions)
				if len(added) > 0 {
					return ctx.RespBadF("缺少权限: %v", added)
				}

				return next(ctx)
			}
		}
	}
}
