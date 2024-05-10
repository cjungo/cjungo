package mid

import (
	"sync"

	"github.com/cjungo/cjungo"
	"github.com/elliotchance/pie/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/exp/constraints"
)

type AuthKeyHandle[T Permission] func(cjungo.HttpContext) ([]T, error)

type Permission interface {
	constraints.Integer | string
}

type PermitManager[T Permission] struct {
	logger *zerolog.Logger
	proofs sync.Map
	handle AuthKeyHandle[T]
}

type PermitManagerProvide[T Permission] func(
	logger *zerolog.Logger,
) (*PermitManager[T], error)

func NewPermitManager[T Permission](handle AuthKeyHandle[T]) PermitManagerProvide[T] {
	return func(logger *zerolog.Logger) (*PermitManager[T], error) {
		manager := &PermitManager[T]{
			logger: logger,
			handle: handle,
		}
		return manager, nil
	}
}

func (manager *PermitManager[T]) Auth(key string, permissions ...T) {
	manager.proofs.Store(key, permissions)
}

func (manager *PermitManager[T]) Permit(permissions ...T) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.(cjungo.HttpContext)
			if owned, err := manager.handle(ctx); err != nil {
				return err
			} else {
				added, _ := pie.Diff(owned, permissions)
				if len(added) > 0 {
					return ctx.RespBadF("缺少权限: %v", added)
				}
			}

			return next(ctx)
		}
	}
}
