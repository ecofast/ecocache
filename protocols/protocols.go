package protocols

import (
	"encoding/binary"
)

const (
	CmSmDif = 127

	CM_REGSVR = 1
	SM_REGSVR = CmSmDif + CM_REGSVR

	CM_DELSVR = 2
	SM_DELSVR = CmSmDif + CM_DELSVR

	CM_REQSVR = 3
	SM_REQSVR = CmSmDif + CM_REQSVR

	CM_PING = 4
	SM_PING = CmSmDif + CM_PING

	CM_GET = 5
	SM_GET = CmSmDif + CM_GET

	CM_SET = 6
	SM_SET = CmSmDif + CM_SET
)

const (
	SizeOfPacketHeadLen   = 4
	SizeOfPacketHeadCmd   = 1
	SizeOfPacketHeadRet   = 1
	SizeOfPacketHeadParam = 2
	SizeOfPacketHead      = SizeOfPacketHeadLen + SizeOfPacketHeadCmd + SizeOfPacketHeadRet + SizeOfPacketHeadParam
)

type Head struct {
	Len   uint32
	Cmd   uint8
	Ret   uint8
	Param uint16
}

type Body struct {
	Content []byte
}

type Packet struct {
	Head
	Body
}

func (self *Packet) Bytes() []byte {
	buf := make([]byte, SizeOfPacketHead+len(self.Body.Content))
	binary.LittleEndian.PutUint32(buf, self.Len)
	buf[SizeOfPacketHeadLen] = self.Cmd
	buf[SizeOfPacketHeadLen+SizeOfPacketHeadCmd] = self.Ret
	binary.LittleEndian.PutUint16(buf[SizeOfPacketHeadLen+SizeOfPacketHeadCmd+SizeOfPacketHeadRet:], self.Param)
	if len(self.Body.Content) > 0 {
		copy(buf[SizeOfPacketHead:], self.Body.Content)
	}
	return buf
}

func NewPacket(cmd, ret uint8, param uint16, content []byte) *Packet {
	p := &Packet{
		Head{
			Len:   uint32(len(content)),
			Cmd:   cmd,
			Ret:   ret,
			Param: param,
		},
		Body{
			Content: nil,
		},
	}
	sz := len(content)
	if sz > 0 {
		p.Body.Content = make([]byte, sz)
		copy(p.Body.Content, content)
	}
	return p
}
