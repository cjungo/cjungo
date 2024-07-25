package ext

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cjungo/cjungo"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type SseEventPair struct {
	Key   string
	Value string
}

type SseEvent struct {
	ID     string
	Event  string
	Data   any
	Others []SseEventPair
}

type SseManager struct {
	logger *zerolog.Logger
}

func NewSseManager(
	logger *zerolog.Logger,
) *SseManager {
	return &SseManager{
		logger: logger,
	}
}

type SseDispatcher interface {
	SseDispatch(ctx cjungo.HttpContext, tx chan SseEvent, rx chan error)
}

func writefln(response *echo.Response, format string, args ...any) error {
	_, err := fmt.Fprintf(response, format, args...)
	return err
}

func (manager *SseManager) Manage(
	dispatcher SseDispatcher,
) cjungo.HttpHandlerFunc {
	return func(ctx cjungo.HttpContext) error {
		reqId := ctx.GetReqID()

		response := ctx.Response()
		response.Header().Set("Content-Type", "text/event-stream")
		response.Header().Set("Cache-Control", "no-cache")
		response.Header().Set("Connection", "keep-alive")

		manager.logger.Info().
			Str("action", "start").
			Str("reqId", reqId).
			Msg("[SSE]")

		tx := make(chan SseEvent)
		rx := make(chan error)
		defer close(rx)
		go func() {
			dispatcher.SseDispatch(ctx, tx, rx)
			close(tx)
		}()
		for {
			select {
			case <-ctx.Request().Context().Done():
				manager.logger.Info().
					Str("action", "done").
					Str("reqId", reqId).
					Msg("[SSE]")
				return nil
			case msg, ok := <-tx:
				// 结束
				if !ok {
					return nil
				}
				manager.logger.Info().
					Str("action", "tx").
					Any("msg", msg).
					Str("reqId", reqId).
					Msg("[SSE]")

				// 错误
				if err, ok := msg.Data.(error); ok {
					return err
				}

				// 消息
				pairs := []SseEventPair{}
				if len(msg.ID) > 0 {
					pairs = append(pairs, SseEventPair{Key: "id", Value: msg.ID})
				} else {
					// time.RFC3339Nano
					pairs = append(pairs, SseEventPair{Key: "id", Value: time.Now().Format("20060102150405.9999")})
				}
				if len(msg.Event) > 0 {
					pairs = append(pairs, SseEventPair{Key: "event", Value: msg.Event})
				}
				if msg.Data != nil {
					data, err := json.Marshal(msg.Data)
					if err != nil {
						return err
					}
					pairs = append(pairs, SseEventPair{
						Key:   "data",
						Value: string(data),
					})
				}
				pairs = append(pairs, msg.Others...)

				for _, pair := range pairs {
					if err := writefln(response, "%s: %s\n", pair.Key, pair.Value); err != nil {
						return err
					}
				}
				if err := writefln(response, "\n"); err != nil {
					return err
				}
				response.Flush()
			}
		}
	}

}
