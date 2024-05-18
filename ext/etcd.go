package ext

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/rs/zerolog"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/dig"
)

type EtcdDiscovery struct {
	client     *clientv3.Client
	serverList sync.Map
	Logger     *zerolog.Logger
}

type EtcdDiscoveryConf struct {
	Endpoints     []string
	DialTimeout   time.Duration
	DefaultPrefix string
}

type EtcdDiscoveryDi struct {
	dig.In
	Logger *zerolog.Logger
	Conf   *EtcdDiscoveryConf `optional:"true"`
}

func NewEtcdDiscovery(di EtcdDiscoveryDi) (*EtcdDiscovery, error) {
	if di.Conf == nil {
		di.Conf = &EtcdDiscoveryConf{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 40 * time.Second,
		}
		di.Logger.Info().Str("action", "启用默认配置").Msg("[ETCD]")
	}

	client, err := clientv3.New(
		clientv3.Config{
			Endpoints:   di.Conf.Endpoints,
			DialTimeout: di.Conf.DialTimeout,
		},
	)
	if err != nil {
		return nil, err
	}

	return &EtcdDiscovery{
		client:     client,
		serverList: sync.Map{},
		Logger:     di.Logger,
	}, nil
}

func (discovery *EtcdDiscovery) WatchService(prefix string) error {
	resp, err := discovery.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		k := string(kv.Key)
		v := string(kv.Value)
		discovery.serverList.Store(k, v)
		discovery.Logger.Info().
			Str("action", "WATCH START").
			Str("service", v).
			Msg("[ETCD]")
	}

	// 启动 watch
	go discovery.watch(prefix)

	return nil
}

func (discovery *EtcdDiscovery) watch(prefix string) {
	watchChan := discovery.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	discovery.Logger.Info().
		Str("action", "WATCH").
		Str("prefix", prefix).
		Msg("[ETCD]")
	for resp := range watchChan {
		for _, e := range resp.Events {
			switch e.Type {
			case mvccpb.PUT:
				discovery.serverList.Store(string(e.Kv.Key), string(e.Kv.Value))
			case mvccpb.DELETE:
				discovery.serverList.Delete(string(e.Kv.Key))
			}
		}
	}
}

func (discovery *EtcdDiscovery) Close() error {
	return discovery.client.Close()
}

func (discovery *EtcdDiscovery) ListService() []string {
	result := []string{}
	discovery.serverList.Range(func(key, value any) bool {
		result = append(result, fmt.Sprint(value))
		return true
	})
	return result
}
