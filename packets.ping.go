package sp2p

type pingReq struct{}
func (t *pingReq) T() byte          { return PingReqT }
func (t *pingReq) String() string   { return PingReqS }
func (t *pingReq) OnHandle(p *SP2p, msg *KMsg) {
	node, err := nodeFromKMsg(msg)
	if err != nil {
		getLog().Error("NodeFromKMsg error", "err", err)
		return
	}
	p.tab.UpdateNode(node)
}
