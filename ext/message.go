package ext

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cjungo/cjungo"
	"github.com/rs/zerolog"
	"golang.org/x/exp/constraints"
	"golang.org/x/net/websocket"
)

type MessageKind = string
type MessageToken interface {
	constraints.Integer | string
}
type MessageControllerProvide[T MessageToken] func(logger *zerolog.Logger) *MessageController[T]
type MessageAuthAccess[T MessageToken] func(ctx cjungo.HttpContext) (T, error)

type MessageCoder[T MessageToken] interface {
	Encode(v *Message[T]) ([]byte, error)
	Decode(v *Message[T], b []byte) error
}

type MessageJsonCoder[T MessageToken] struct{}

func (coder *MessageJsonCoder[T]) Encode(v *Message[T]) ([]byte, error) {
	return json.Marshal(v)
}

func (coder *MessageJsonCoder[T]) Decode(v *Message[T], b []byte) error {
	return json.Unmarshal(b, &v)
}

const (
	MESSAGE_AUTH_TOKEN_HEADER             = "X-Message-Auth-Token"
	MESSAGE_SINGLE            MessageKind = "SINGLE"
	MESSAGE_GROUP             MessageKind = "GROUP"
	MESSAGE_ACK               MessageKind = "ACK"
)

type Message[T MessageToken] struct {
	ID     string      `json:"id"`
	Kind   MessageKind `json:"kind"`
	TimeAt time.Time   `json:"timeAt"`
	To     T           `json:"to,omitempty"`
	Group  T           `json:"group,omitempty"`
	From   T           `json:"from,omitempty"`
	Data   any         `json:"data,omitempty"`
}

type MessageClient[T MessageToken] struct {
	Token T
	Conn  *websocket.Conn
}

type MessageController[T MessageToken] struct {
	logger      *zerolog.Logger
	clients     sync.Map
	groups      sync.Map
	tokenAccess MessageAuthAccess[T]
	coder       MessageCoder[T]
}

type MessageControllerProviderConf[T MessageToken] struct {
	TokenAccess MessageAuthAccess[T]
	Coder       MessageCoder[T]
}

func ProvideMessageController[T MessageToken](
	conf *MessageControllerProviderConf[T],
) (MessageControllerProvide[T], error) {
	if conf.TokenAccess == nil {
		return nil, fmt.Errorf("TokenAccess 不可空")
	}
	coder := conf.Coder
	if coder == nil {
		coder = &MessageJsonCoder[T]{}
	}

	return func(
		logger *zerolog.Logger,
	) *MessageController[T] {
		return &MessageController[T]{
			logger:      logger,
			clients:     sync.Map{},
			groups:      sync.Map{},
			tokenAccess: conf.TokenAccess,
			coder:       coder,
		}
	}, nil
}

func MessageJsonEncode() {

}

func (controller *MessageController[T]) Dispatch(ctx cjungo.HttpContext) error {
	token, err := controller.tokenAccess(ctx)
	if err != nil {
		return err
	}
	controller.logger.Info().
		Str("action", "start").
		Any("token", token).
		Msg("[MESSAGE]")
	v, ok := controller.clients.Load(token)
	if ok {
		client := v.(*MessageClient[T])
		if err := client.Conn.Close(); err != nil {
			controller.logger.Error().
				Str("action", "断开旧链接").
				Any("token", token).
				Err(err).
				Msg("[MESSAGE]")
		} else {
			controller.logger.Info().
				Str("action", "断开旧链接").
				Any("token", token).
				Msg("[MESSAGE]")
		}
	}

	errChan := make(chan error, 1)
	websocket.Handler(func(conn *websocket.Conn) {
		client := MessageClient[T]{
			Token: token,
			Conn:  conn,
		}
		controller.clients.Store(token, &client)
		defer func() {
			controller.clients.Delete(token)
			conn.Close()
		}()

		controller.logger.Info().
			Str("action", "open").
			Any("token", token).
			Msg("[MESSAGE]")

		for {
			data := []byte{}
			if err := websocket.Message.Receive(conn, &data); err != nil {
				errChan <- err
				return
			}

			controller.logger.Info().
				Str("action", "receive").
				Any("token", token).
				Str("content", string(data)).
				Msg("[MESSAGE]")

			msg := Message[T]{}
			if err := controller.coder.Decode(&msg, data); err != nil {
				errChan <- err
				return
			}

			controller.logger.Info().
				Str("action", "send").
				Any("token", token).
				Str("kind", msg.Kind).
				Msg("[MESSAGE]")

			switch msg.Kind {
			case MESSAGE_GROUP:
				if err := controller.sendGroup(&client, &msg); err != nil {
					errChan <- err
					return
				}
			default:
				if err := controller.sendSingle(&client, &msg); err != nil {
					data, err := controller.coder.Encode(&Message[T]{
						ID:   msg.ID,
						Kind: MESSAGE_ACK,
						Data: err.Error(),
					})
					if err != nil {
						errChan <- err
						return
					}
					if err2 := websocket.Message.Send(client.Conn, data); err2 != nil {
						errChan <- err2
						return
					}
				}
			}
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())
	err = <-errChan
	controller.logger.Info().
		Str("action", "end").
		Err(err).
		Any("token", token).
		Msg("[MESSAGE]")
	return err
}

func (controller *MessageController[T]) sendSingle(from *MessageClient[T], msg *Message[T]) error {
	t, ok := controller.clients.Load(msg.To)
	if !ok {
		return fmt.Errorf("无效的目标: %v", msg.To)
	}
	target := t.(*MessageClient[T])
	response := Message[T]{
		From: from.Token,
	}
	MoveField(msg, &response)
	data, err := controller.coder.Encode(&response)
	if err != nil {
		return err
	}
	controller.logger.Info().
		Str("action", "singleSend").
		Any("data", string(data)).
		Msg("[MESSAGE]")
	return websocket.Message.Send(target.Conn, data)
}

func (controller *MessageController[T]) sendGroup(from *MessageClient[T], msg *Message[T]) error {
	g, ok := controller.groups.Load(msg.Group)
	if !ok {
		return fmt.Errorf("无效的组: %v", msg.To)
	}
	group := g.([]string)
	for _, tid := range group {
		response := Message[T]{
			From: from.Token,
		}
		MoveField(msg, &response)
		t, ok := controller.clients.Load(tid)
		if !ok {
			controller.logger.Error().
				Str("action", "sendGroup").
				Str("msg", "无效组员ID").
				Str("tid", tid).
				Msg("[MESSAGE]")
		} else {
			target := t.(*MessageClient[T])
			data, err := controller.coder.Encode(&response)
			if err != nil {
				return err
			}
			if err := websocket.Message.Send(target.Conn, data); err != nil {
				return err
			}
		}
	}
	return nil
}
