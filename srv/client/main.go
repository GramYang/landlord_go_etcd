package main

import (
	"bufio"
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/timer"
	"landlord_go/proto"
	"landlord_go/service"
	"os"
	"strings"
	"time"
)

type ClientParam struct {
	NetPeerType string
	NetProcName string
}

func login(param *ClientParam) (agentAddr, gameSvcID string) {
	log.Debugln("Create login connection...")
	loginPeer := CreateConnection("login", param.NetPeerType, param.NetProcName)
	_ = RemoteCall(loginPeer.(cellnet.Session), &proto.LoginREQ{
		Version:  "1.0",
		Platform: "demo",
		TokenReq: []byte("1234"),
	}, func(ack *proto.LoginACK) {
		if ack.Result == proto.ResultCode_NoError {
			agentAddr = fmt.Sprintf("%s:%d", ack.Server.IP, ack.Server.Port)
			gameSvcID = ack.GameSvcID
		} else {
			panic(ack.Result.String())
		}
		//实现短连接
		loginPeer.Stop()
	})
	return
}

func getAgentSession(agentAddr string, param *ClientParam) (ret cellnet.Session) {
	log.Debugln("Prepare agent connection...")
	waitGameReady := make(chan struct{})
	go KeepConnection("agent", agentAddr, param.NetPeerType, param.NetProcName, func(ses cellnet.Session) {
		ret = ses
		waitGameReady <- struct{}{}
	}, func() {
		os.Exit(0)
	})
	<-waitGameReady
	log.Debugln("Agent connection ready")
	return
}

func ReadConsole(callback func(string)) {
	for {
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			break
		}
		text = strings.TrimSpace(text)
		callback(text)
	}
}

func main() {
	service.Init("client")
	service.ConnectDiscovery()
	var currParam *ClientParam
	currParam = &ClientParam{NetPeerType:"tcp.SyncConnector", NetProcName:"tcp.demo"}
	agentAddr, gameSvcID := login(currParam)
	if agentAddr == "" {
		return
	}
	fmt.Println("agent:", agentAddr)
	agentSes := getAgentSession(agentAddr, currParam)
	_ = RemoteCall(agentSes, &proto.VerifyREQ{
		GameToken:"verify",
		GameSvcID:gameSvcID,
	}, func(ack *proto.VerifyACK) {
		fmt.Println(ack)
	})
	timer.NewLoop(nil, time.Second * 5, func(loop *timer.Loop) {
		agentSes.Send(&proto.PingACK{})
	}, nil).Start()
	fmt.Println("Start chat now !")
	ReadConsole(func(s string) {
		//_ = RemoteCall(agentSes, &proto.ChatREQ{
		//	Content:s,
		//}, func(ack *proto.ChatACK) {
		//	fmt.Println(ack.Content)
		//})
	})
}