package protocols

import (
	"encoding/binary"
)

const (
	cCmSmDif = 127

	CM_REGSVR = 1
	SM_REGSVR = cCmSmDif + CM_REGSVR

	CM_DELSVR = 2
	SM_DELSVR = cCmSmDif + CM_DELSVR

	CM_REQSVR = 3
	SM_REQSVR = cCmSmDif + CM_REQSVR

	CM_PING = 4
	SM_PING = cCmSmDif + CM_PING

	CM_GET = 5
	SM_GET = cCmSmDif + CM_GET

	CM_MGET = 6
	SM_MGET = cCmSmDif + CM_MGET
)

const (
	SizeOfPacketHeadLen      = 4
	SizeOfPacketHeadCmd      = 1
	SizeOfPacketHeadReserved = 1
	SizeOfPacketHeadParam    = 2
	SizeOfPacketHead         = SizeOfPacketHeadLen + SizeOfPacketHeadCmd + SizeOfPacketHeadReserved + SizeOfPacketHeadParam
)

type Head struct {
	Len   uint32
	Cmd   uint8
	_     uint8 // Reserved
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
	// buf[SizeOfPacketHeadLen+SizeOfPacketHeadCmd] = self.Reserved
	binary.LittleEndian.PutUint16(buf[SizeOfPacketHeadLen+SizeOfPacketHeadCmd+SizeOfPacketHeadReserved:], self.Param)
	if len(self.Body.Content) > 0 {
		copy(buf[SizeOfPacketHead:], self.Body.Content)
	}
	return buf
}

func NewPacket(cmd uint8, param uint16, content []byte) *Packet {
	p := &Packet{
		Head{
			Len:   uint32(len(content)),
			Cmd:   cmd,
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
