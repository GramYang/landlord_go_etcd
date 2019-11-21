package etcdv3

import (
	"github.com/coreos/etcd/clientv3"
	"time"
)

func DefaultConfig() clientv3.Config{
	return clientv3.Config{
		Endpoints:[]string{"localhost:2379",},
		DialTimeout: 10 * time.Second,
	}
}