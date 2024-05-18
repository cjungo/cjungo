package ext

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/rs/zerolog"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/dig"
)

type EtcdPair struct {
	Key   string
	Value string
}

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
			Str("key", k).
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

func (discovery *EtcdDiscovery) ListService() []EtcdPair {
	result := []EtcdPair{}
	discovery.serverList.Range(func(key, value any) bool {
		result = append(result, EtcdPair{
			Key:   fmt.Sprint(key),
			Value: fmt.Sprint(value),
		})
		return true
	})
	return result
}

func (discovery *EtcdDiscovery) FindService(prefix string) EtcdPair {
	var result EtcdPair
	discovery.serverList.Range(func(key, value any) bool {
		path := fmt.Sprint(key)
		if strings.HasPrefix(path, prefix) {
			result = EtcdPair{
				Key:   path,
				Value: fmt.Sprint(value),
			}
			return false
		}
		return true
	})
	return result
}

// ====================================================
type EtcdRegister struct {
	client *clientv3.Client
	Logger *zerolog.Logger
}

type EtcdRegisterConf struct {
	Endpoints   []string
	DialTimeout time.Duration
}

type EtcdRegisterDi struct {
	dig.In
	Logger *zerolog.Logger
	Conf   *EtcdRegisterConf `optional:"true"`
}

func NewEtcdRegister(di EtcdRegisterDi) (*EtcdRegister, error) {
	if di.Conf == nil {
		di.Logger.Info().
			Str("tip", "启用默认配置").
			Msg("[ETCD]")
		di.Conf = &EtcdRegisterConf{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 40 * time.Second,
		}
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

	return &EtcdRegister{
		client: client,
		Logger: di.Logger,
	}, nil
}

type EtcdLeasePair struct {
	EtcdPair
	KeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	LeaseTtl      int64
	LeaseID       clientv3.LeaseID
}

func (register *EtcdRegister) RegisterPair(pair EtcdPair, ttl int64) (*EtcdLeasePair, error) {
	lease, err := register.client.Grant(context.Background(), ttl)
	if err != nil {
		return nil, err
	}
	if _, err := register.client.Put(context.Background(), pair.Key, pair.Value, clientv3.WithLease(lease.ID)); err != nil {
		return nil, err
	}
	leaseChan, err := register.client.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		return nil, err
	}
	return &EtcdLeasePair{
		EtcdPair:      pair,
		KeepAliveChan: leaseChan,
		LeaseTtl:      ttl,
		LeaseID:       lease.ID,
	}, nil
}
