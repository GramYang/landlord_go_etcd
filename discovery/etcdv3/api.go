package etcdv3

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"landlord_go/discovery"
	"sync"
)

type notifyContext struct {
	mode  string
}

type etcdv3Discovery struct {
	client *clientv3.Client

	kvCache map[string]string
	kvCacheGuard sync.RWMutex

	svcCache map[string][]*discovery.ServiceDesc //key是svcName，svcName标记的是服务类型，服务用svcID标记
	svcCacheGuard sync.RWMutex

	notifyMap sync.Map
}

func NewDiscovery(config interface{}) discovery.Discovery {
	if config == nil {
		config = DefaultConfig()
	}
	cli, err := clientv3.New(config.(clientv3.Config))
	if err != nil {
		log.Fatalf("create clientv3 failed, err: %s\n", err.Error())
	}
	self := &etcdv3Discovery{
		client:cli,
		kvCache:make(map[string]string),
		svcCache:make(map[string][]*discovery.ServiceDesc),
	}
	go self.watch()
	return self
}

func (self *etcdv3Discovery) Register(des *discovery.ServiceDesc) error {
	if des.Name == "" {
		return errors.New("invalid svc name")
	}
	if des.ID == "" {
		return errors.New("invalid svc id")
	}
	data, err := json.Marshal(des)
	if err != nil {
		return err
	}
	key := discovery.ServiceKeyPrefix + des.ID
	lea, _ := self.client.Grant(context.TODO(), 5)
	_, err = self.client.Put(context.TODO(), key, string(data), clientv3.WithLease(lea.ID))
	if err == nil {
		log.Tracef("register key: %s, value: %s \n", key, string(data))
	}
	_, err = self.client.KeepAlive(context.TODO(), lea.ID)
	return err
}

func (self *etcdv3Discovery) Deregister(svcid string) error {
	if svcid != "" {
		key := discovery.ServiceKeyPrefix+svcid
		_, err := self.client.Delete(context.TODO(), key)
		log.Tracef("deregister key: %s on the etcd\n", key)
		return err
	} else {
		return errors.New("service deregister invalid params")
	}
}

func (self *etcdv3Discovery) Query(name string) (ret []*discovery.ServiceDesc) {
	//这里出现了错误，svcCache是svcid，而name是key前缀加svcname。都是没有和etcd适配好的锅
	self.svcCacheGuard.RLock()
	ret = self.svcCache[name]
	self.svcCacheGuard.RUnlock()
	if ret != nil {
		return ret
	} else {
		res, err := self.client.Get(context.TODO(), name, clientv3.WithPrefix())
		if err != nil {
			return nil
		} else {
			if len(res.Kvs) > 0 {
				var r *discovery.ServiceDesc
				err = json.Unmarshal(res.Kvs[0].Value, r)
				if err != nil {
					log.Errorln(err)
					return nil
				}
				ret = append(ret, r)
				return ret
			}
			return nil
		}
	}
}

func (self *etcdv3Discovery) RegisterNotify(mode string) (ret chan struct{}) {
	ret = make(chan struct{}, 10)
	switch mode {
	case "add":
		self.notifyMap.Store(ret, &notifyContext{
			mode:mode,
		})
	default:
		panic("unknown notify mode: " + mode)
	}
	return
}

func (self *etcdv3Discovery) DeregisterNotify(mode string, c chan struct{}) {
	switch mode {
	case "add":
		self.notifyMap.Store(c, nil)
	default:
		panic("unknown notify mode: " + mode)
	}
}

func (self *etcdv3Discovery) SetValue(key string, value interface{}) error {
	valueStr := discovery.AnyToString(value)
	_, err := self.client.Put(context.TODO(), discovery.KVKeyPrefix + key, valueStr)
	return err
}

func (self *etcdv3Discovery) GetValue(key string) (string, error) {
	res, err := self.client.Get(context.TODO(), discovery.KVKeyPrefix + key)
	if err != nil {
		return "", err
	} else {
		if len(res.Kvs) != 0 {
			return string(res.Kvs[0].Value), err
		}
		return "", nil
	}
}

func (self *etcdv3Discovery) DeleteValue(key string) error {
	_, err := self.client.Delete(context.TODO(), discovery.KVKeyPrefix + key)
	return err
}

//增加一个Close方法，因为etcdv3并不是基于cellnet的，不能统一关闭
func (self *etcdv3Discovery) Close() error{
	return self.client.Close()
}