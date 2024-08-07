package ext

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cjungo/cjungo"
	"github.com/rs/zerolog"
	"golang.org/x/net/websocket"
)

type MessageKind = string
type MessageToken = any
type MessageControllerProvide[T MessageToken] func(logger *zerolog.Logger) (*MessageController[T], error)
type MessageAuthAccess[T MessageToken] func(ctx cjungo.HttpContext) (T, error)
type OnMessageRecv[T MessageToken] func(controller *MessageController[T], client *MessageClient[T], msg *Message[T]) error

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

func (client *MessageClient[T]) Call(coder MessageCoder[T], msg *Message[T]) error {
	data, err := coder.Encode(msg)
	if err != nil {
		return err
	}
	return websocket.Message.Send(client.Conn, data)
}

func (client *MessageClient[T]) Recv(coder MessageCoder[T], msg *Message[T]) error {
	data := []byte{}
	if err := websocket.Message.Receive(client.Conn, &data); err != nil {
		return err
	}
	return coder.Decode(msg, data)
}

type MessageController[T MessageToken] struct {
	logger      *zerolog.Logger
	clients     sync.Map
	groups      sync.Map
	tokenAccess MessageAuthAccess[T]
	coder       MessageCoder[T]
	onRecv      OnMessageRecv[T]
}

type MessageControllerProviderConf[T MessageToken] struct {
	TokenAccess MessageAuthAccess[T]
	Coder       MessageCoder[T]
	OnRecv      OnMessageRecv[T]
}

func ProvideMessageController[T MessageToken](
	conf *MessageControllerProviderConf[T],
) MessageControllerProvide[T] {
	coder := conf.Coder
	if coder == nil {
		coder = &MessageJsonCoder[T]{}
	}
	onRecv := conf.OnRecv
	if onRecv == nil {
		onRecv = defaultOnRecv
	}

	return func(
		logger *zerolog.Logger,
	) (*MessageController[T], error) {
		if conf.TokenAccess == nil {
			return nil, fmt.Errorf("TokenAccess 不可空")
		}
		return &MessageController[T]{
			logger:      logger,
			clients:     sync.Map{},
			groups:      sync.Map{},
			tokenAccess: conf.TokenAccess,
			coder:       coder,
			onRecv:      onRecv,
		}, nil
	}
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

		if err := controller.handle(&client); err != nil {
			errChan <- err
			return
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

func (controller *MessageController[T]) FindClient(token T) (*MessageClient[T], error) {
	t, ok := controller.clients.Load(token)
	if !ok {
		return nil, fmt.Errorf("invalid MessageClient token: %v", token)
	}
	return t.(*MessageClient[T]), nil
}

func (controller *MessageController[T]) FindGroup(group T) ([]T, error) {
	g, ok := controller.groups.Load(group)
	if !ok {
		return nil, fmt.Errorf("invalid MessageClient Group token: %v", group)
	}
	return g.([]T), nil
}

func (controller *MessageController[T]) handle(client *MessageClient[T]) error {
	for {
		msg := Message[T]{}
		if err := client.Recv(controller.coder, &msg); err != nil {
			return err
		}
		if err := controller.onRecv(controller, client, &msg); err != nil {
			return err
		}
	}
}

func (controller *MessageController[T]) sendSingle(from *MessageClient[T], msg *Message[T]) error {
	target, err := controller.FindClient(msg.To)
	if err != nil {
		return err
	}

	response := Message[T]{
		From: from.Token,
	}
	MoveField(msg, &response)
	return target.Call(controller.coder, &response)
}

func (controller *MessageController[T]) sendGroup(from *MessageClient[T], msg *Message[T]) error {
	group, err := controller.FindGroup(msg.Group)
	if err != nil {
		return err
	}
	for _, tid := range group {
		target, err := controller.FindClient(tid)
		if err != nil {
			controller.logger.Error().
				Str("action", "sendGroup").
				Str("msg", "无效组员ID").
				Any("tid", tid).
				Msg("[MESSAGE]")
		} else {
			response := Message[T]{
				From: from.Token,
			}
			MoveField(msg, &response)
			if err := target.Call(controller.coder, &response); err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultOnRecv[T MessageToken](
	controller *MessageController[T],
	client *MessageClient[T],
	msg *Message[T],
) error {
	controller.logger.Info().
		Str("action", "send").
		Any("token", client.Token).
		Str("kind", msg.Kind).
		Msg("[MESSAGE]")

	switch msg.Kind {
	case MESSAGE_GROUP:
		if err := controller.sendGroup(client, msg); err != nil {
			return err
		}
	default:
		if err := controller.sendSingle(client, msg); err != nil {
			if err := client.Call(controller.coder, &Message[T]{
				ID:   msg.ID,
				Kind: MESSAGE_ACK,
				Data: err.Error(),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
