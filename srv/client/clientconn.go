package main

import (
	"errors"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/gorillaws"
	"github.com/davyxu/cellnet/proc/tcp"
	"landlord_go/discovery"
	"reflect"
	"sync"
	"time"
)

var (
	callByType sync.Map // map[reflect.Type]func(interface{})
)

func selectStrategy(descList []*discovery.ServiceDesc) *discovery.ServiceDesc {
	if len(descList) == 0 {
		return nil
	}

	return descList[0]
}

func queryServiceAddress(serviceName string) (*discovery.ServiceDesc, error) {
	descList := discovery.Default.Query(serviceName)
	desc := selectStrategy(descList)
	if desc == nil {
		return nil, errors.New("target not reachable:" + serviceName)
	}
	return desc, nil
}

func CreateConnection(serviceName, netPeerType, netProcName string) (ret cellnet.Peer) {
	//每一次add事件就会触发一次循环
	notify := discovery.Default.RegisterNotify("add")
	done := make(chan struct{})
	go func() {
		for {
			desc, err := queryServiceAddress(serviceName)
			if err == nil {
				p := peer.NewGenericPeer(netPeerType, serviceName, desc.Address(), nil)
				proc.BindProcessorHandler(p, netProcName, nil)
				p.Start()
				conn := p.(cellnet.PeerReadyChecker)
				if conn.IsReady() {
					ret = p
					break
				}
				p.Stop()
			}
			<-notify
		}
		discovery.Default.DeregisterNotify("add", notify)
		done <- struct{}{}
	}()
	<-done
	return
}

func RemoteCall(target, req , callback interface{}) error {
	funcType := reflect.TypeOf(callback)
	if funcType.Kind() != reflect.Func {
		panic("callback require 'func'")
	}
	var ses cellnet.Session
	switch tgt := target.(type) {
	case cellnet.Session:
		ses = tgt
	default:
		panic("rpc: Invalid peer type, require cellnet.Session")
	}
	if ses == nil {
		return errors.New("Empty session")
	}
	feedBack := make(chan interface{})
	ackType := funcType.In(0)
	if funcType.NumIn() != 1 || ackType.Kind() != reflect.Ptr {
		panic("callback func param format like 'func(ack *YouMsgAck)'")
	}
	ackType = ackType.Elem()
	callByType.Store(ackType, feedBack)
	defer callByType.Delete(ackType)
	ses.Send(req)
	select {
	case ack := <-feedBack:
		vCall := reflect.ValueOf(callback)
		vCall.Call([]reflect.Value{reflect.ValueOf(ack)})
		return nil
	case <-time.After(time.Second):
		log.Errorln("RemoteCall: RPC time out")
		return errors.New("RPC Time out")
	}
	return nil
}

//保持长连接
func KeepConnection(svcid, addr, netPeerType, netProc string, onReady func(cellnet.Session), onClose func()) {
	var stop sync.WaitGroup
	p := peer.NewGenericPeer(netPeerType, svcid, addr, nil)
	proc.BindProcessorHandler(p, netProc, func(ev cellnet.Event) {
		switch ev.Message().(type) {
		case *cellnet.SessionClosed:
			stop.Done()
		}
	})
	stop.Add(1)
	p.Start()
	conn := p.(cellnet.PeerReadyChecker)
	if conn.IsReady() {
		onReady(p.(cellnet.TCPConnector).Session())
		stop.Wait()
	}
	p.Stop()
	if onClose != nil {
		onClose()
	}
}

type TypeRPCHooker struct {
}

func (TypeRPCHooker) OnInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	outputEvent, _, err := ResolveInboundEvent(inputEvent)
	if err != nil {
		log.Errorln("rpc.ResolveInboundEvent", err)
		return
	}
	return outputEvent
}

func (TypeRPCHooker) OnOutboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event) {
	return inputEvent
}

func ResolveInboundEvent(inputEvent cellnet.Event) (outputEvent cellnet.Event, handled bool, err error) {
	incomingMsgType := reflect.TypeOf(inputEvent.Message()).Elem()
	if rawFeedback, ok := callByType.Load(incomingMsgType); ok {
		feedBack := rawFeedback.(chan interface{})
		feedBack <- inputEvent.Message()
		return inputEvent, true, nil
	}
	return inputEvent, false, nil
}

func init() {
	proc.RegisterProcessor("tcp.demo", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(tcp.TCPMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(new(tcp.MsgHooker), new(TypeRPCHooker)))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})

	proc.RegisterProcessor("ws.demo", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(gorillaws.WSMessageTransmitter))
		bundle.SetHooker(proc.NewMultiHooker(new(gorillaws.MsgHooker), new(TypeRPCHooker)))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
}