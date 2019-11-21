package service

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/util"
	"landlord_go/discovery"
	"landlord_go/discovery/etcdv3"
	"os"
	"os/signal"
	"syscall"
)

type peerListener interface {
	Port() int
}

type ServiceMeta map[string]string

type DiscoveryOption struct {
	MaxCount      int    // 连接数，默认发起多条连接，只有在大于0时才有效
	MatchSvcGroup string // 空时，匹配所有同类服务，否则找指定组的服务
}

//将来会进行其他功能的扩展
func Init(name string) {
	svcName = name
}

func ConnectDiscovery() {
	//一般不会用分布式服务发现
	log.Traceln("Connecting to discovery ...")
	discovery.Default = etcdv3.NewDiscovery(nil)
}

func WaitExitSignal() {
	log.Traceln("Waiting for exit signal...")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
}

func Register(p cellnet.Peer, options ...interface{}) *discovery.ServiceDesc {
	host := util.GetLocalIP()
	property := p.(cellnet.PeerProperty)
	sd := &discovery.ServiceDesc{
		ID:MakeLocalSvcID(property.Name()),
		Name:property.Name(),
		Host:host,
		Port:p.(peerListener).Port(),
	}
	sd.SetMeta("SvcGroup", "dev")
	sd.SetMeta("SvcIndex", "0")
	for _, opt := range options {
		switch optValue := opt.(type) {
		case ServiceMeta:
			for  metaKey, metaValue := range optValue {
				sd.SetMeta(metaKey, metaValue)
			}
		}
	}
	log.Debugf("service '%s' listen at port: %d", sd.ID, sd.Port)
	p.(cellnet.ContextSet).SetContext("sd", sd)
	_ = discovery.Default.Deregister(sd.ID)
	 err := discovery.Default.Register(sd)
	if err != nil {
		log.Errorf("service register failed, %s %s", sd.String(), err.Error())
	}
	return sd
}

func Unregister(p cellnet.Peer) {
	property := p.(cellnet.PeerProperty)
	_ = discovery.Default.Deregister(MakeLocalSvcID(property.Name()))
}

//发现一种服务并创建connector连接服务，因为这种服务会有多个地址，所有会创建多个connector，由multiPeer持有
func DiscoveryService(tgtSvcName string, opt DiscoveryOption, peerCreator func(MultiPeer, *discovery.ServiceDesc)) cellnet.Peer {
	multiPeer := NewMultiPeer()
	go func() {
		//保证一个add事件创建一个连接
		notify := discovery.Default.RegisterNotify("add")
		for {
			QueryService(tgtSvcName,
				Filter_MatchSvcGroup(opt.MatchSvcGroup),
				func(desc *discovery.ServiceDesc) interface{} {
					prePeer := multiPeer.GetPeer(desc.ID)
					if prePeer != nil {
						var preDesc *discovery.ServiceDesc
						if prePeer.(cellnet.ContextSet).FetchContext("sd", &preDesc) && !preDesc.Equals(desc) {
							log.Infof("service '%s' change desc, %+v -> %+v...", desc.ID, preDesc, desc)
							multiPeer.RemovePeer(desc.ID)
							prePeer.Stop()
						} else {
							return true
						}
					}
					if opt.MaxCount > 0 && len(multiPeer.GetPeers()) >= opt.MaxCount {
						return true
					}
					peerCreator(multiPeer, desc)
					return true
				})
			<-notify
		}
	}()
	return multiPeer
}

func Close() error {
	log.Traceln("close etcd clientv3...")
	return discovery.Default.Close()
}