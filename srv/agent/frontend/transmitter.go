package frontend

import (
	"encoding/binary"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/util"
	"github.com/gorilla/websocket"
	"io"
	"net"
)

type socketOpt interface {
	MaxPacketSize() int
	ApplySocketReadTimeout(conn net.Conn, callback func())
	ApplySocketWriteTimeout(conn net.Conn, callback func())
}

type directTCPTransmitter struct {
}

func (directTCPTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {
	reader, ok := ses.Raw().(io.Reader)
	if !ok || reader == nil {
		return nil, nil
	}
	opt := ses.Peer().(socketOpt)
	if conn, ok := ses.Raw().(net.Conn); ok {
		for {
			opt.ApplySocketReadTimeout(conn, func() {
				var msgID int
				var msgData []byte
				msgID, msgData, err = RecvLTVPacketData(reader, opt.MaxPacketSize())
				if err == nil {
					msg, err = ProcFrontendPacket(msgID, msgData, ses)
				}
			})
			if err != nil {
				break
			}
		}
	}
	return
}

func (directTCPTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) (err error) {
	writer, ok := ses.Raw().(io.Writer)
	if !ok || writer == nil {
		return nil
	}
	opt := ses.Peer().(socketOpt)
	opt.ApplySocketWriteTimeout(writer.(net.Conn), func() {
		err = util.SendLTVPacket(writer, ses.(cellnet.ContextSet), msg)
	})
	return
}

type directWSMessageTransmitter struct{
}

func (directWSMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {
	conn, ok := ses.Raw().(*websocket.Conn)
	if !ok || conn == nil {
		return nil, nil
	}
	var (
		messageType int
		raw []byte
	)
	for {
		messageType, raw, err = conn.ReadMessage()
		if err != nil {
			break
		}
		switch messageType {
		case websocket.BinaryMessage:
			msgID := binary.LittleEndian.Uint16(raw)
			msgData := raw[msgIDSize:]
			msg, err = ProcFrontendPacket(int(msgID), msgData, ses)
		}
		if err != nil {
			break
		}
	}
	return
}

func (directWSMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {
	conn, ok := ses.Raw().(*websocket.Conn)
	if !ok || conn == nil {
		return nil
	}
	var (
		msgID int
		msgData []byte
	)
	switch m := msg.(type) {
	case *cellnet.RawPacket:
		msgID = m.MsgID
		msgData = m.MsgData
	default:
		var err error
		var meta *cellnet.MessageMeta
		msgData, meta, err = codec.EncodeMessage(msg, nil)
		if err != nil {
			return err
		}
		msgID = meta.ID
	}
	pkt := make([]byte, msgIDSize + len(msgData))
	binary.LittleEndian.PutUint16(pkt, uint16(msgID))
	copy(pkt[msgIDSize:], msgData)
	_ = conn.WriteMessage(websocket.BinaryMessage, pkt)
	return nil
}