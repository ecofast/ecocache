package msgnode

type MsgNode struct {
	Consumer chan<- *MsgNode
	Cmd      uint8
	_        uint8
	Param    uint16
	_        uint32
	Body     []byte
	Next     *MsgNode
}

func New(consumer chan<- *MsgNode, cmd uint8, param uint16, body []byte) *MsgNode {
	return &MsgNode{
		Consumer: consumer,
		Cmd:      cmd,
		Param:    param,
		Body:     body,
	}
}
