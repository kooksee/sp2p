package sp2p

type pingReq struct{}

func (t *pingReq) T() byte        { return pingReqT }
func (t *pingReq) String() string { return pingReqS }
func (t *pingReq) OnHandle(p ISP2P, msg *KMsg) {
	node, err := nodeFromKMsg(msg)
	if err != nil {
		getLog().Error("NodeFromKMsg error", "err", err)
		return
	}
	p.UpdateNode(node.string())
}
