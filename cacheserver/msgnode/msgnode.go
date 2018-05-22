package msgnode

type MsgNode struct {
	Consumer chan<- *MsgNode
	Cmd      uint8
	Ret      uint8
	Param    uint16
	_        uint32
	Body     []byte
	Next     *MsgNode
}

func New(consumer chan<- *MsgNode, cmd, ret uint8, param uint16, body []byte) *MsgNode {
	return &MsgNode{
		Consumer: consumer,
		Cmd:      cmd,
		Ret:      ret,
		Param:    param,
		Body:     body,
	}
}
