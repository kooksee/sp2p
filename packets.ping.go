package sp2p

type PingReq struct{}
func (t *PingReq) T() byte          { return PingReqT }
func (t *PingReq) String() string   { return PingReqS }
func (t *PingReq) Create() IMessage { return &PingReq{} }
func (t *PingReq) OnHandle(p *SP2p, msg *KMsg) {
	node, err := NodeFromKMsg(msg)
	if err != nil {
		logger.Error("NodeFromKMsg error", "err", err)
		return
	}
	p.tab.UpdateNode(node)
}
