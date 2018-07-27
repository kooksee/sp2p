package sp2p

type pingReq struct {
	IMessage
}

func (t *pingReq) T() byte        { return pingReqT }
func (t *pingReq) String() string { return pingReqS }
func (t *pingReq) OnHandle(p ISP2P, msg *KMsg) error {
	return p.NodeUpdate(msg.FN)
}
