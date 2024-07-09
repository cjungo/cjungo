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

type PermitProof[TP Permission, TS any] interface {
	GetPermissions() []TP
	GetStore() TS
}

type AuthKeyHandle[TP Permission, TS any] func(cjungo.HttpContext) (PermitProof[TP, TS], error)

type PermitManager[TP Permission, TS any] struct {
	logger *zerolog.Logger
	proofs sync.Map
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

func (manager *PermitManager[TP, TS]) GetProof(ctx cjungo.HttpContext) (PermitProof[TP, TS], bool) {
	if v, ok := manager.proofs.Load(ctx.GetReqID()); ok {
		return v.(PermitProof[TP, TS]), ok
	}
	return nil, false
}

func permit[TP Permission, TS any](proof PermitProof[TP, TS], permissions ...TP) bool {
	intersect := pie.Intersect(proof.GetPermissions(), permissions)
	return len(intersect) > 0
}

// 权限逻辑 OR ，如果要实现 AND ，连续使用 多个 Permit 中间件
func (manager *PermitManager[TP, TS]) Permit(permissions ...TP) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.(cjungo.HttpContext)
			reqID := ctx.GetReqID()
			manager.logger.Info().Any("reqID", reqID).Msg("[PermitManager]")

			// 当多个 Permit 中间件合用时，此处获取其他 Permit 中间件初始化好的。
			if pp, ok := manager.proofs.Load(reqID); ok {
				if permit(pp.(PermitProof[TP, TS]), permissions...) {
					return next(ctx)
				}
				return ctx.RespBadF("缺少权限: %v", permissions)
			}

			if pp, err := manager.handle(ctx); err != nil {
				return err
			} else {
				// 当多个 Permit 中间件合用时，此处做初始化。
				manager.proofs.Store(reqID, pp)
				manager.logger.Info().Any("store", pp).Str("id", reqID).Msg("[PermitManager]")
				defer func() {
					manager.logger.Info().Any("delete", pp).Str("id", reqID).Msg("[PermitManager]")
					manager.proofs.Delete(reqID)
				}()

				if permit(pp, permissions...) {
					return next(ctx)
				}
				return ctx.RespBadF("缺少权限: %v", permissions)
			}
		}
	}
}
