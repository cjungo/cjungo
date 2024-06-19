package cjungo

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type TaskStatus string

const (
	TASK_STATUS_PENDING          TaskStatus = "Pending"
	TASK_STATUS_PROCESSING       TaskStatus = "Processing"
	TASK_STATUS_START            TaskStatus = "Start"
	TASK_STATUS_OK               TaskStatus = "Ok"
	TASK_STATUS_FAILED           TaskStatus = "Failed"
	TASK_STATUS_NOT_HAVE_PROCESS TaskStatus = "Not have process"
)

type TaskConfig struct {
}

type TaskResult struct {
	ID     string
	Name   string
	Status TaskStatus
	Data   TaskResultMessage
}

type TaskResultMessage map[string]any
type TaskActionParam map[string]any
type TaskActionProcess func(param *TaskAction) (TaskResultMessage, error)

type TaskAction struct {
	ID    string
	Name  string
	Param TaskActionParam
}

type TaskQueue struct {
	Logger      *zerolog.Logger
	unprocessed chan *TaskAction
	processes   sync.Map
	results     sync.Map
}

type TaskQueueDi struct {
	dig.In
	Conf   *TaskConfig `optional:"true"`
	Logger *zerolog.Logger
}
type TaskQueueProvide func(di TaskQueueDi) (*TaskQueue, error)

func NewTaskQueueHandle(initialize func(*TaskQueue) error) TaskQueueProvide {
	return func(di TaskQueueDi) (*TaskQueue, error) {
		queue := &TaskQueue{
			Logger:      di.Logger,
			unprocessed: make(chan *TaskAction, 1),
			processes:   sync.Map{},
			results:     sync.Map{},
		}

		err := initialize(queue)
		return queue, err
	}
}

func (queue *TaskQueue) setStatus(action *TaskAction, status TaskStatus) {
	if r, ok := queue.results.Load(action.ID); ok {
		r.(*TaskResult).Status = status
		queue.results.Store(action.ID, r)
	} else {
		queue.Logger.Error().Str("name", action.Name).Str("id", action.ID).Str("action", "没有该任务的状态信息").Msg("[TASK]")
	}
}

func (queue *TaskQueue) setData(action *TaskAction, data TaskResultMessage) {
	if r, ok := queue.results.Load(action.ID); ok {
		r.(*TaskResult).Data = data
		queue.results.Store(action.ID, r)
	} else {
		queue.Logger.Error().Str("name", action.Name).Str("id", action.ID).Str("action", "没有该任务的状态信息").Msg("[TASK]")
	}
}

func (queue *TaskQueue) Run() error {
	go func() {
		queue.Logger.Info().Str("action", "队列启动").Msg("[TASK]")

		for action := range queue.unprocessed {

			process, ok := queue.processes.Load(action.Name)

			if !ok {
				queue.Logger.Error().Str("name", action.Name).Str("id", action.ID).Str("action", "没有该类型的处理器").Msg("[TASK]")
				queue.setStatus(action, TASK_STATUS_NOT_HAVE_PROCESS)
				continue
			}

			queue.setStatus(action, TASK_STATUS_START)
			data, err := process.(TaskActionProcess)(action)
			queue.setData(action, data)
			if err != nil {
				queue.setStatus(action, TASK_STATUS_FAILED)
				queue.Logger.Error().
					Str("action", "任务处理出错").
					Str("name", action.Name).
					Str("id", action.ID).
					Any("result", data).
					AnErr("error", err).
					Msg("[TASK]")
			}
			queue.setStatus(action, TASK_STATUS_OK)
			queue.Logger.Info().
				Str("action", "完成任务").
				Str("name", action.Name).
				Str("id", action.ID).
				Any("result", data).
				Msg("[TASK]")
		}

		queue.Logger.Info().Str("action", "队列关闭").Msg("[TASK]")
	}()
	return nil
}

func (queue *TaskQueue) RegisterProcess(name string, process TaskActionProcess) {
	queue.processes.Store(name, process)
}

func (queue *TaskQueue) PushTask(name string, param TaskActionParam) (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	action := &TaskAction{
		ID:    id.String(),
		Name:  name,
		Param: param,
	}
	result := &TaskResult{
		ID:     action.ID,
		Name:   action.Name,
		Status: TASK_STATUS_PENDING,
	}
	queue.results.Store(action.ID, result)
	queue.unprocessed <- action
	return id.String(), nil
}

func (queue *TaskQueue) QueryTask(id string) (*TaskResult, error) {
	if result, ok := queue.results.Load(id); ok {
		return result.(*TaskResult), nil
	}
	return nil, fmt.Errorf("没有 ID：%s 的队列信息", id)
}

func LoadTaskConfFromEnv(logger *zerolog.Logger) (*TaskConfig, error) {
	logger.Info().Str("action", "通过环境变量配置任务队列").Msg("[TASK]")
	conf := &TaskConfig{}
	return conf, nil
}
