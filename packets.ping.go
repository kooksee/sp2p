package sp2p

type PingReq struct {
	t byte
	s string
}

func (t *PingReq) T() byte        { return t.t }
func (t *PingReq) String() string { return t.s }
func (t *PingReq) OnHandle(p *SP2p, msg *KMsg) {
	node, err := NodeFromKMsg(msg)
	if err != nil {
		logger.Error("NodeFromKMsg error", "err", err)
		return
	}

	p.tab.UpdateNode(node)
}
