package frontend

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	bodySize = 2
	msgIDSize = 2
)

func RecvLTVPacketData(reader io.Reader, maxPacketSize int) (msgID int, msgData []byte, err error) {
	var sizeBuffer = make([]byte, bodySize)
	//先读2个字节
	_, err = io.ReadFull(reader, sizeBuffer)
	if err != nil {
		return
	}
	if len(sizeBuffer) < bodySize {
		err = errors.New("packet short size")
		return
	}
	//统一用小端格式读取size
	size := binary.LittleEndian.Uint16(sizeBuffer)
	if maxPacketSize > 0 && size >= uint16(maxPacketSize) {
		err = errors.New("packet over size")
		return
	}
	body := make([]byte, size)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return
	}
	if len(body) < bodySize {
		err = errors.New("short msgid")
		return
	}
	msgID = int(binary.LittleEndian.Uint16(body))
	msgData = body[msgIDSize:]
	return
}