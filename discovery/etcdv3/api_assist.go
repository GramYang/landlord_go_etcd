package etcdv3

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"landlord_go/discovery"
	"time"
)

func (self *etcdv3Discovery) watch() {
	serviceCh := self.client.Watch(context.TODO(), discovery.ServiceKeyPrefix, clientv3.WithPrefix())
	for res1 := range serviceCh {
		for _, ev := range res1.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				desc := &discovery.ServiceDesc{}
				if err := json.Unmarshal([]byte(ev.Kv.Value), desc); err != nil {
					log.Errorf("ServiceDesc unmarshal failed, %s", err)
				}
				if discovery.IsServiceKey(string(ev.Kv.Key)) {
					self.updateSvcCache(desc.ID, desc)
				}
			case clientv3.EventTypeDelete:
				if SvcID := discovery.GetSvcIDByServiceKey(string(ev.Kv.Key)); SvcID != "" {
					self.deleteSvcCache(SvcID)
				}
			}
		}
	}
	kvCh := self.client.Watch(context.TODO(), discovery.KVKeyPrefix, clientv3.WithPrefix())
	for res2 := range kvCh {
		for _, ev := range res2.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				if discovery.IsKVKey(string(ev.Kv.Key)) {
					self.updateKVCache(string(ev.Kv.Key), string(ev.Kv.Value))
				}
			case clientv3.EventTypeDelete:
				if key := discovery.GetKVKey(string(ev.Kv.Key)); key != "" {
					self.deleteKVCache(key)
				}
			}
		}
	}
}

func (self *etcdv3Discovery) updateSvcCache(svcID string, desc *discovery.ServiceDesc) {
	self.svcCacheGuard.Lock()
	svcName := discovery.GetSvcNameByID(svcID)
	list := self.svcCache[svcName]
	var notFound = true
	for index, svc := range list {
		if svc.ID == desc.ID {
			list[index] = desc
			notFound = false
			break
		}
	}
	if notFound {
		list = append(list, desc)
	}
	self.svcCache[svcName] = list
	self.svcCacheGuard.Unlock()
	self.triggerNotify("add", time.Second * 10)
}

func (self *etcdv3Discovery) deleteSvcCache(svcID string) {
	svcName := discovery.GetSvcNameByID(svcID)
	list := self.svcCache[svcName]
	for index, svc := range list {
		if svc.ID == svcID {
			list = append(list[:index], list[index + 1:]...)
			break
		}
	}
	self.svcCache[svcName] = list
}

func (self *etcdv3Discovery) updateKVCache(key, value string) {
	self.kvCacheGuard.Lock()
	self.kvCache[key] = value
	self.kvCacheGuard.Unlock()
}

func (self *etcdv3Discovery) deleteKVCache(key string) {
	self.kvCacheGuard.Lock()
	delete(self.kvCache, key)
	self.kvCacheGuard.Unlock()
}

func (self *etcdv3Discovery) triggerNotify(mode string, timeout time.Duration) {
	self.notifyMap.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		ctx := value.(*notifyContext)
		if ctx.mode != mode {
			return true
		}
		c := key.(chan struct{})
		if timeout == 0 {
			select {
			case c <- struct{}{}:
			default: //阻塞了什么都不做
			}
		} else {
			select {
			case c <- struct{}{}:
			case <- time.After(timeout):
				// 接收通知阻塞太久，或者没有解注册的channel
				log.Errorf("notify(%s) timeout, not free? regstack: %s", ctx.mode)
			}
		}
		return true
	})
}